// Package config provides the configuration for tests run as part of the osde2e suite.
package config

const (
	// EnvVarTag is the Go struct tag containing the environment variable that sets the option.
	EnvVarTag = "env"

	// SectionTag is the Go struct tag containing the documentation section of the option.
	SectionTag = "sect"

	// DefaultTag is the Go struct tag containing the default value of the option.
	DefaultTag = "default"
)

// Cfg is the configuration used for end to end testing.
var Cfg = new(Config)

// Config dictates the behavior of cluster tests.
type Config struct {
	// ReportDir is the location JUnit XML results are written.
	ReportDir string `env:"REPORT_DIR" sect:"tests"`

	// Suffix is used at the end of test names to identify them.
	Suffix string `env:"SUFFIX" sect:"tests"`

	// DryRun lets you run osde2e all the way up to the e2e tests then skips them.
	DryRun bool `env:"DRY_RUN" sect:"tests"`

	// OCMToken is used to authenticate with OCM.
	OCMToken string `env:"OCM_TOKEN" sect:"required"`

	// JobName is the name of the job that is running osde2e.
	JobName string `env:"JOB_NAME" sect:"tests"`

	// ClusterID identifies the cluster. If set at start, an existing cluster is tested.
	ClusterID string `env:"CLUSTER_ID" sect:"cluster"`

	// ClusterName is the name of the cluster being created.
	ClusterName string `env:"CLUSTER_NAME" sect:"cluster"`

	// ClusterVersion is the version of the cluster being deployed.
	ClusterVersion string `env:"CLUSTER_VERSION" sect:"version"`

	// MajorTarget is the major version to target. If specified, it is used in version selection.
	MajorTarget int64 `env:"MAJOR_TARGET" sect:"version"`

	// MinorTarget is the minor version to target. If specified, it is used in version selection.
	MinorTarget int64 `env:"MINOR_TARGET" sect:"version"`

	// PollingTimeout is how long (in mimutes) to wait for an object to be created
	// before failing the test.
	PollingTimeout int64 `env:"POLLING_TIMEOUT" sect:"tests" default:"30"`

	// MultiAZ deploys a cluster across multiple availability zones.
	MultiAZ bool `env:"MULTI_AZ" sect:"cluster" default:"false"`

	// DestroyClusterAfterTest set to false if you want OCM to clean up the cluster itself after the test completes.
	DestroyClusterAfterTest bool `env:"DESTROY_CLUSTER" sect:"cluster" default:"true"`

	// Kubeconfig is used to access a cluster.
	Kubeconfig []byte `env:"TEST_KUBECONFIG" sect:"cluster"`

	// ClusterExpiryInMinutes is how long before a cluster expires and is deleted by OSD.
	ClusterExpiryInMinutes int64 `env:"CLUSTER_EXPIRY_IN_MINUTES" sect:"cluster" default:"210"`

	// AfterTestClusterWait is how long to keep a cluster around after tests have run.
	AfterTestClusterWait int64 `env:"AFTER_TEST_CLUSTER_WAIT" sect:"environment" default:"60"`

	// ClusterUpTimeout is how long to wait before failing a cluster launch.
	ClusterUpTimeout int64 `env:"CLUSTER_UP_TIMEOUT" sect:"environment" default:"135"`

	// OSDEnv is the OpenShift Dedicated environment used to provision clusters.
	OSDEnv string `env:"OSD_ENV" sect:"environment" default:"prod"`

	// DebugOSD shows debug level messages when enabled.
	DebugOSD bool `env:"DEBUG_OSD" sect:"environment" default:"false"`

	// GinkgoSkip is a regex passed to Ginkgo that skips any test suites matching the regex. ex. "Operator"
	GinkgoSkip string `env:"GINKGO_SKIP" sect:"tests"`

	// GinkgoFocus is a regex passed to Ginkgo that focus on any test suites matching the regex. ex. "Operator"
	GinkgoFocus string `env:"GINKGO_FOCUS" sect:"tests"`

	// CleanRuns is the number of times the test-version is run before skipping.
	CleanRuns int `env:"CLEAN_RUNS" sect:"tests"`

	// UpgradeReleaseStream used to retrieve latest release images. If set, it will be used to perform an upgrade.
	UpgradeReleaseStream string `env:"UPGRADE_RELEASE_STREAM" sect:"upgrade"`

	// UpgradeReleaseName is the name of the release in a release stream. UpgradeReleaseStream must be set.
	UpgradeReleaseName string `env:"UPGRADE_RELEASE_NAME" sect:"upgrade"`

	// UpgradeImage is the release image a cluster is upgraded to. If set, it overrides the release stream and upgrades.
	UpgradeImage string `env:"UPGRADE_IMAGE" sect:"upgrade"`

	// OperatorSkip is a comma-delimited list of operator names to ignore health checks from. ex. "insights,telemetry"
	OperatorSkip string `env:"OPERATOR_SKIP" sect:"tests" default:"insights"`

	// UploadMetrics tells osde2e whether to try to upload to the S3 metrics bucket.
	UploadMetrics bool `env:"UPLOAD_METRICS" sect:"metrics" default:"false"`

	// MetricsBucket is the bucket that metrics data will be uploaded to.
	MetricsBucket string `env:"METRICS_BUCKET" sect:"metrics" default:"osde2e-metrics"`

	// InstalledWorkloads is an internal variable used to track currently installed workloads in this test run.
	InstalledWorkloads map[string]string

	// Phase is an internal variable used to track the current set of tests being run (install, upgrade).
	Phase string
}
