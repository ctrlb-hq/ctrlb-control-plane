package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/otelcol"
)

func TestCollectorBootstrapConfig(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "config", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := dryRunCollectorConfig(t, string(data)); err != nil {
		t.Fatalf("bundled VM bootstrap configuration is incompatible: %v", err)
	}
}

// The backend exports this fixture in its own module. Supplying its path runs
// compatibility validation without coupling the two modules at build time.
func TestGeneratedKubernetesPresetCompatibility(t *testing.T) {
	path := os.Getenv("KUBERNETES_PRESET_FIXTURE")
	if path == "" {
		t.Skip("set KUBERNETES_PRESET_FIXTURE to the generated backend preset")
	}
	t.Setenv("STATE_DIR", t.TempDir())
	t.Setenv("K8S_NODE_NAME", "worker-1")
	t.Setenv("K8S_NODE_IP", "127.0.0.1")
	t.Setenv("K8S_CLUSTER_NAME", "test-cluster")
	t.Setenv("POD_NAME", "ctrlb-collector-test")
	t.Setenv("POD_NAMESPACE", "observability")
	t.Setenv("POD_UID", "test-pod-uid")
	provider, err := otelcol.NewConfigProvider(collectorConfigProviderSettings(path))
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Shutdown(context.Background())
	factories, err := componentsFactory()
	if err != nil {
		t.Fatal(err)
	}
	config, err := provider.Get(context.Background(), factories)
	if err != nil {
		t.Fatal(err)
	}
	if err := confmap.Validate(config); err != nil {
		t.Fatalf("generated service-account preset failed typed validation: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var hostConfig map[string]any
	if err := json.Unmarshal(data, &hostConfig); err != nil {
		t.Fatal(err)
	}
	for name, receiver := range hostConfig["receivers"].(map[string]any) {
		if strings.HasPrefix(name, "kubelet_stats/") {
			receiver.(map[string]any)["auth_type"] = "none"
		}
	}
	data, err = json.Marshal(hostConfig)
	if err != nil {
		t.Fatal(err)
	}
	if err := dryRunCollectorConfig(t, string(data)); err != nil {
		t.Fatalf("generated preset failed host DryRun with only kubelet auth changed: %v", err)
	}
}

func TestCollectorLegacyAWSMetricsConfig(t *testing.T) {
	const config = `
receivers:
  awscloudwatchmetrics:
    region: us-east-1
    metrics:
      named:
        - namespace: AWS/EC2
          metric_name: CPUUtilization
          period: 5m
          aws_aggregation: Average
exporters:
  debug: {}
service:
  pipelines:
    metrics:
      receivers: [awscloudwatchmetrics]
      exporters: [debug]
`
	if err := dryRunCollectorConfig(t, config); err != nil {
		t.Fatalf("legacy AWS metrics configuration is incompatible: %v", err)
	}
}

func TestCollectorLegacyVMConfig(t *testing.T) {
	const config = `
receivers:
  hostmetrics:
    collection_interval: 30s
    scrapers:
      cpu: {}
      memory: {}
processors:
  batch: {}
exporters:
  otlphttp:
    endpoint: http://127.0.0.1:4318
service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: [batch]
      exporters: [otlphttp]
`
	if err := dryRunCollectorConfig(t, config); err != nil {
		t.Fatalf("legacy VM component names are incompatible: %v", err)
	}
}

func TestCollectorKubernetesConfig(t *testing.T) {
	// DryRun constructs the kubelet client. Use its unauthenticated mode here so
	// this compatibility test does not require in-cluster service account files.
	t.Setenv("CTRL_TEST_KUBELET_ENDPOINT", "http://127.0.0.1:10255")
	t.Setenv("CTRL_TEST_NODE_NAME", "test-node")
	t.Setenv("CTRL_TEST_CLUSTER_NAME", "test-cluster")

	config := fmt.Sprintf(`
extensions:
  file_storage:
    directory: %q
    create_directory: true
  health_check: {}
receivers:
  filelog:
    include: [/var/log/pods/*/*/*.log]
    exclude: [/var/log/pods/*/ctrlb-collector/*.log]
    include_file_path: true
    start_at: end
    storage: file_storage
    operators:
      - type: container
        id: container-parser
  kubeletstats:
    collection_interval: 30s
    auth_type: none
    endpoint: ${env:CTRL_TEST_KUBELET_ENDPOINT}
    node: ${CTRL_TEST_NODE_NAME}
    metric_groups: [node, pod, container]
processors:
  memory_limiter:
    check_interval: 1s
    limit_mib: 128
  k8sattributes:
    auth_type: serviceAccount
    filter:
      node_from_env_var: CTRL_TEST_NODE_NAME
    extract:
      metadata: [k8s.namespace.name, k8s.pod.name, k8s.pod.uid, k8s.node.name, k8s.container.name]
    pod_association:
      - sources:
          - from: resource_attribute
            name: k8s.pod.uid
  resource:
    attributes:
      - key: k8s.cluster.name
        value: ${env:CTRL_TEST_CLUSTER_NAME}
        action: upsert
  batch: {}
exporters:
  debug: {}
service:
  extensions: [file_storage, health_check]
  pipelines:
    logs:
      receivers: [filelog]
      processors: [memory_limiter, k8sattributes, resource, batch]
      exporters: [debug]
    metrics:
      receivers: [kubeletstats]
      processors: [memory_limiter, k8sattributes, resource, batch]
      exporters: [debug]
`, filepath.Join(t.TempDir(), "offsets"))

	if err := dryRunCollectorConfig(t, config); err != nil {
		t.Fatalf("Kubernetes collection configuration is incompatible: %v", err)
	}

	invalid := strings.Replace(config, "collection_interval: 30s", "collection_interval: -1s", 1)
	if err := dryRunCollectorConfig(t, invalid); err == nil {
		t.Fatal("invalid kubelet configuration passed Collector validation")
	}
}

func dryRunCollectorConfig(t *testing.T, config string) error {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "collector.yaml")
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	previousPath := constants.AGENT_CONFIG_PATH
	constants.AGENT_CONFIG_PATH = configPath
	defer func() { constants.AGENT_CONFIG_PATH = previousPath }()

	collector, err := getNewOTELCollector()
	if err != nil {
		return err
	}
	return collector.DryRun(context.Background())
}
