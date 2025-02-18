// Package common launches an OSD cluster, performs tests on it, and destroys it.
package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/helper"
	"github.com/openshift/osde2e/pkg/metadata"
	"github.com/openshift/osde2e/pkg/osd"
	"github.com/openshift/osde2e/pkg/upgrade"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// CustomMetadataFile is the name of the custom metadata file generated for spyglass visualization.
	CustomMetadataFile string = "custom-prow-metadata.json"
)

// OSD is used to deploy and manage clusters.
var OSD *osd.OSD

// RunE2ETests runs the osde2e test suite using the given cfg.
func RunE2ETests(t *testing.T, cfg *config.Config) {
	var err error
	gomega.RegisterFailHandler(ginkgo.Fail)

	// set defaults
	if cfg.Suffix == "" {
		cfg.Suffix = randomStr(3)
	}

	if cfg.ReportDir == "" {
		if dir, err := ioutil.TempDir("", "osde2e"); err == nil {
			cfg.ReportDir = dir
		}
	}

	// setup OSD unless Kubeconfig is present
	if len(cfg.Kubeconfig) > 0 {
		log.Print("Found an existing Kubeconfig!")
	} else {
		if OSD, err = osd.New(cfg.OCMToken, cfg.OSDEnv, cfg.DebugOSD); err != nil {
			t.Fatalf("could not setup OSD: %v", err)
		}

		metadata.Instance.Environment = cfg.OSDEnv

		// check that enough quota exists for this test if creating cluster
		if len(cfg.ClusterID) == 0 {
			if enoughQuota, err := OSD.CheckQuota(cfg); err != nil {
				log.Printf("Failed to check if enough quota is available: %v", err)
			} else if !enoughQuota {
				t.Fatal("Currently not enough quota exists to run this test, failing...")
			}
		}

		// configure cluster and upgrade versions
		if err = ChooseVersions(cfg, OSD); err != nil {
			t.Fatalf("failed to configure versions: %v", err)
		}
	}

	// setup reporter
	if err = os.Mkdir(cfg.ReportDir, os.ModePerm); err != nil {
		log.Printf("Could not create reporter directory: %v", err)
	}

	if !cfg.DryRun {
		log.Println("Running e2e tests...")

		runTestsInPhase(t, cfg, "install", "OSD e2e suite")

		// upgrade cluster if requested
		if cfg.UpgradeImage != "" || cfg.UpgradeReleaseStream != "" {
			if cfg.Kubeconfig != nil {
				if err = upgrade.RunUpgrade(cfg, OSD); err != nil {
					t.Errorf("error performing upgrade: %v", err)
				}

				log.Println("Running e2e tests POST-UPGRADE...")
				runTestsInPhase(t, cfg, "upgrade", "OSD e2e suite post-upgrade")
			} else {
				log.Println("No Kubeconfig found from initial cluster setup. Unable to run upgrade.")
			}
		}

		if cfg.ReportDir != "" {
			if err = metadata.Instance.WriteToJSON(filepath.Join(cfg.ReportDir, CustomMetadataFile)); err != nil {
				t.Errorf("error while writing metadata: %v", err)
			}

			prometheusFilename, err := NewMetrics(cfg).WritePrometheusFile(cfg.ReportDir)
			if err != nil {
				t.Errorf("error while writing prometheus metrics: %v", err)
			}

			if cfg.UploadMetrics {
				if err := uploadFileToMetricsBucket(cfg, filepath.Join(cfg.ReportDir, prometheusFilename)); err != nil {
					t.Errorf("error while uploading prometheus metrics: %v", err)
				}
			}
		}

		if OSD != nil {
			if cfg.DestroyClusterAfterTest {
				log.Printf("Destroying cluster '%s'...", cfg.ClusterID)
				if err = OSD.DeleteCluster(cfg.ClusterID); err != nil {
					t.Errorf("error deleting cluster: %v", err)
				}
			} else {
				log.Printf("For debugging, please look for cluster ID %s in environment %s", cfg.ClusterID, cfg.OSDEnv)
			}
		} else {
			// If we run against an arbitrary cluster and not a ci-specific cluster
			// we need to clean up our workload tests manually.
			h := &helper.H{
				Config: cfg,
			}
			h.SetupNoProj()

			log.Printf("Cleaning up workloads tests")
			workloads := h.GetWorkloads()
			for _, project := range workloads {
				log.Printf("Deleting Project: %s", project)
				h.SetProjectByName(project)
				h.Project().ProjectV1().Projects().Delete(project, &metav1.DeleteOptions{})
			}
		}
	}
}

func runTestsInPhase(t *testing.T, cfg *config.Config, phase string, description string) {
	cfg.Phase = phase
	phaseDirectory := filepath.Join(cfg.ReportDir, phase)
	if _, err := os.Stat(phaseDirectory); os.IsNotExist(err) {
		if err := os.Mkdir(phaseDirectory, os.FileMode(0755)); err != nil {
			t.Fatalf("error while creating phase directory %s", phaseDirectory)
		}
	}
	phaseReportPath := filepath.Join(phaseDirectory, fmt.Sprintf("junit_%v.xml", cfg.Suffix))
	phaseReporter := reporters.NewJUnitReporter(phaseReportPath)
	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, description, []ginkgo.Reporter{phaseReporter})
}

// uploadFileToMetricsBucket uploads the given file (with absolute path) to the metrics S3 bucket "incoming" directory.
func uploadFileToMetricsBucket(cfg *config.Config, filename string) error {
	// We're very intentionally using the shared configs here.
	// This allows us to configure the AWS client at a system level and this should behave as expected.
	// This is particularly useful if we want to, at some point in the future, run this on an AWS host with an instance profile
	// that doesn't need explicit credentials.
	session, err := session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable})
	if err != nil {
		return err
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	uploader := s3manager.NewUploader(session)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(cfg.MetricsBucket),
		Key:    aws.String(path.Join("incoming", filepath.Base(filename))),
		Body:   file,
	})

	return err
}
