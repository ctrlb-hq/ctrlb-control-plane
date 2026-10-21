package kubernetes

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
)

type CreatePipelineRequest struct {
	Name                      string            `json:"name"`
	Endpoint                  string            `json:"endpoint"`
	LogsEndpoint              string            `json:"logs_endpoint,omitempty"`
	MetricsEndpoint           string            `json:"metrics_endpoint,omitempty"`
	Headers                   map[string]string `json:"headers,omitempty"`
	KubeletInsecureSkipVerify bool              `json:"kubelet_insecure_skip_verify,omitempty"`
}

type CreatePipelineResponse struct {
	PipelineID int `json:"pipeline_id"`
}

// PipelineGraph is the editable Kubernetes preset. Its environment placeholders
// resolve on each collector, so one desired revision applies to every node.
func PipelineGraph(req CreatePipelineRequest) (models.PipelineGraph, error) {
	if strings.TrimSpace(req.Name) == "" || len(req.Name) > 253 {
		return models.PipelineGraph{}, fmt.Errorf("%w: a pipeline name of at most 253 characters is required", ErrInvalid)
	}
	for index, value := range []string{req.Endpoint, req.LogsEndpoint, req.MetricsEndpoint} {
		if index > 0 && value == "" {
			continue
		}
		endpoint, err := url.Parse(value)
		if err != nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.Hostname() == "" || endpoint.User != nil || endpoint.Fragment != "" {
			return models.PipelineGraph{}, fmt.Errorf("%w: endpoints must be absolute HTTP or HTTPS URLs without embedded credentials or a fragment", ErrInvalid)
		}
	}
	for name, value := range req.Headers {
		if strings.TrimSpace(name) == "" || strings.ContainsAny(name+value, "\r\n") {
			return models.PipelineGraph{}, fmt.Errorf("%w: header names must be nonempty and headers cannot contain line breaks", ErrInvalid)
		}
	}
	headers := make(map[string]string, len(req.Headers))
	for key, value := range req.Headers {
		// Collector config resolution treats dollars as expansion syntax. The
		// setup API accepts literal credentials, including dollar characters.
		headers[key] = strings.ReplaceAll(value, "$", "$$")
	}
	signals := []string{"logs", "metrics"}
	node := func(id int, name, component, role string, supported []string, config map[string]any) models.PipelineNodes {
		return models.PipelineNodes{ComponentID: id, Name: name, ComponentName: component, ComponentRole: role, SupportedSignals: supported, Config: config}
	}
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			node(1, "Pod logs", "filelog_receiver", "receiver", []string{"logs"}, map[string]any{
				"include":  []string{"/var/log/pods/*/*/*.log"},
				"exclude":  []string{"/var/log/pods/${env:POD_NAMESPACE}_${env:POD_NAME}_${env:POD_UID}/*/*.log"},
				"start_at": "end", "include_file_path": true, "storage": "file_storage",
				"operators": []any{map[string]any{"type": "container", "id": "container-parser", "add_metadata_from_filepath": true}},
			}),
			node(2, "Node pod and container metrics", "kubeletstats_receiver", "receiver", []string{"metrics"}, map[string]any{
				"collection_interval": "30s", "auth_type": "serviceAccount", "endpoint": "https://${env:K8S_NODE_IP}:10250",
				"node": "${env:K8S_NODE_NAME}", "ca_file": "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
				"insecure_skip_verify": req.KubeletInsecureSkipVerify, "metric_groups": []string{"node", "pod", "container"},
			}),
			node(3, "Memory limit", "memorylimiter_processor", "processor", signals, map[string]any{
				"check_interval": "1s", "limit_mib": 400, "spike_limit_mib": 80,
			}),
			node(4, "Kubernetes metadata", "k8sattributes_processor", "processor", signals, map[string]any{
				"auth_type": "serviceAccount", "passthrough": false,
				"filter":  map[string]any{"node_from_env_var": "K8S_NODE_NAME"},
				"extract": map[string]any{"metadata": []string{"k8s.namespace.name", "k8s.pod.name", "k8s.pod.uid", "k8s.node.name", "k8s.container.name"}},
				"pod_association": []any{
					map[string]any{"sources": []any{map[string]any{"from": "resource_attribute", "name": "k8s.pod.uid"}}},
					map[string]any{"sources": []any{map[string]any{"from": "resource_attribute", "name": "k8s.pod.ip"}}},
				},
			}),
			node(5, "Cluster and node", "resource_processor", "processor", signals, map[string]any{
				"attributes": []any{
					map[string]any{"key": "k8s.cluster.name", "value": "${env:K8S_CLUSTER_NAME}", "action": "upsert"},
					map[string]any{"key": "k8s.node.name", "value": "${env:K8S_NODE_NAME}", "action": "upsert"},
				},
			}),
			node(6, "Batch", "batch_processor", "processor", signals, map[string]any{"timeout": "5s", "send_batch_size": 1024}),
			node(7, "CtrlB ingestion", "otlphttp_exporter", "exporter", signals, map[string]any{
				"endpoint": req.Endpoint, "headers": headers, "compression": "none", "encoding": "proto", "timeout": "30s",
				"retry_on_failure": map[string]any{"enabled": true, "initial_interval": "1s", "max_interval": "10s", "max_elapsed_time": "300s"},
				"sending_queue":    map[string]any{"enabled": true, "queue_size": 2048, "num_consumers": 2},
			}),
		},
		Edges: []models.PipelineEdges{{Source: "1", Target: "3"}, {Source: "2", Target: "3"}, {Source: "3", Target: "4"}, {Source: "4", Target: "5"}, {Source: "5", Target: "6"}, {Source: "6", Target: "7"}},
	}
	if req.LogsEndpoint != "" {
		graph.Nodes[len(graph.Nodes)-1].Config["logs_endpoint"] = req.LogsEndpoint
	}
	if req.MetricsEndpoint != "" {
		graph.Nodes[len(graph.Nodes)-1].Config["metrics_endpoint"] = req.MetricsEndpoint
	}
	return graph, nil
}

func (s *Service) CreatePipeline(ctx context.Context, req CreatePipelineRequest, createdBy string) (*CreatePipelineResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	graph, err := PipelineGraph(req)
	if err != nil {
		return nil, err
	}
	id, err := frontendpipeline.NewFrontendPipelineRepository(s.repository.db).CreatePipeline(models.CreatePipelineRequest{
		Name: strings.TrimSpace(req.Name), CreatedBy: createdBy, PipelineGraph: graph,
	})
	if err != nil {
		return nil, err
	}
	pipelineID, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	return &CreatePipelineResponse{PipelineID: pipelineID}, nil
}
