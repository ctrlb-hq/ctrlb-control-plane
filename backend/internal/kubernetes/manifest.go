package kubernetes

import (
	"bytes"
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultCollectorImage = "ghcr.io/ctrlb-hq/ctrlb-otel-collector:latest"

type ManifestRequest struct {
	Token      string `json:"token"`
	BackendURL string `json:"backend_url"`
	Image      string `json:"image"`
}

var installationIDPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)
var imagePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._:/@-]*$`)

// Manifest authenticates the installation token before returning its Secret.
// Encoding structured documents keeps user values out of YAML syntax.
func (s *Service) Manifest(ctx context.Context, id string, req ManifestRequest) ([]byte, string, error) {
	if !installationIDPattern.MatchString(id) {
		return nil, "", ErrUnauthorized
	}
	i, err := s.repository.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", ErrUnauthorized
	}
	if err != nil {
		return nil, "", err
	}
	if req.Token == "" || subtle.ConstantTimeCompare([]byte(tokenHash(req.Token)), []byte(i.tokenHash)) != 1 {
		return nil, "", ErrUnauthorized
	}
	u, err := url.Parse(req.BackendURL)
	if err != nil || len(req.BackendURL) > 2048 || strings.TrimSpace(req.BackendURL) != req.BackendURL || strings.ContainsAny(req.BackendURL, "\r\n\\#") ||
		(u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" {
		return nil, "", fmt.Errorf("%w: backend_url must be an absolute HTTPS or explicitly selected HTTP URL without credentials, query or fragment", ErrInvalid)
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return nil, "", fmt.Errorf("%w: backend_url has an invalid port", ErrInvalid)
		}
	}
	if req.Image == "" {
		req.Image = DefaultCollectorImage
	}
	if len(req.Image) > 512 || !imagePattern.MatchString(req.Image) {
		return nil, "", fmt.Errorf("%w: image must be a container image reference", ErrInvalid)
	}
	name := "ctrlb-collector-" + id[:12]
	labels := map[string]any{"app.kubernetes.io/name": "ctrlb-collector", "app.kubernetes.io/instance": name, "ctrlb.ai/installation-id": id}
	metadata := func(namespace bool) map[string]any {
		m := map[string]any{"name": name, "labels": labels}
		if namespace {
			m["namespace"] = "observability"
		}
		return m
	}
	fieldEnv := func(name, path string) map[string]any {
		return map[string]any{"name": name, "valueFrom": map[string]any{"fieldRef": map[string]any{"fieldPath": path}}}
	}
	env := []any{
		map[string]any{"name": "INSTALLATION_ID", "value": id},
		map[string]any{"name": "BACKEND_URL", "value": strings.TrimRight(req.BackendURL, "/")},
		map[string]any{"name": "INSTALLATION_TOKEN_FILE", "value": "/var/run/secrets/ctrlb/token"},
		map[string]any{"name": "K8S_CLUSTER_NAME", "value": i.ClusterName},
		map[string]any{"name": "STATE_DIR", "value": "/var/lib/ctrlb"},
		fieldEnv("K8S_NODE_NAME", "spec.nodeName"), fieldEnv("K8S_NODE_IP", "status.hostIP"),
		fieldEnv("POD_NAMESPACE", "metadata.namespace"), fieldEnv("POD_NAME", "metadata.name"),
		fieldEnv("POD_UID", "metadata.uid"), fieldEnv("POD_IP", "status.podIP"),
	}
	if u.Scheme == "http" {
		env = append(env, map[string]any{"name": "ALLOW_INSECURE_HTTP", "value": "true"})
	}
	probe := func(path string, failures int) map[string]any {
		return map[string]any{"httpGet": map[string]any{"path": path, "port": "management"}, "periodSeconds": 5, "timeoutSeconds": 2, "failureThreshold": failures}
	}
	documents := []any{
		map[string]any{"apiVersion": "v1", "kind": "Namespace", "metadata": map[string]any{"name": "observability"}},
		map[string]any{"apiVersion": "v1", "kind": "Secret", "metadata": metadata(true), "type": "Opaque", "stringData": map[string]any{"token": req.Token}},
		map[string]any{"apiVersion": "v1", "kind": "ServiceAccount", "metadata": metadata(true)},
		map[string]any{"apiVersion": "rbac.authorization.k8s.io/v1", "kind": "ClusterRole", "metadata": metadata(false), "rules": []any{
			map[string]any{"apiGroups": []string{""}, "resources": []string{"pods", "namespaces"}, "verbs": []string{"get", "list", "watch"}},
			map[string]any{"apiGroups": []string{""}, "resources": []string{"nodes/stats"}, "verbs": []string{"get"}},
		}},
		map[string]any{"apiVersion": "rbac.authorization.k8s.io/v1", "kind": "ClusterRoleBinding", "metadata": metadata(false),
			"roleRef":  map[string]any{"apiGroup": "rbac.authorization.k8s.io", "kind": "ClusterRole", "name": name},
			"subjects": []any{map[string]any{"kind": "ServiceAccount", "name": name, "namespace": "observability"}}},
		map[string]any{"apiVersion": "apps/v1", "kind": "DaemonSet", "metadata": metadata(true), "spec": map[string]any{
			"selector":       map[string]any{"matchLabels": map[string]any{"app.kubernetes.io/instance": name}},
			"updateStrategy": map[string]any{"type": "RollingUpdate", "rollingUpdate": map[string]any{"maxUnavailable": 1}},
			"template": map[string]any{"metadata": map[string]any{"labels": labels}, "spec": map[string]any{
				"serviceAccountName": name, "nodeSelector": map[string]any{"kubernetes.io/os": "linux"},
				"terminationGracePeriodSeconds": 45,
				"tolerations":                   []any{map[string]any{"key": "node-role.kubernetes.io/control-plane", "operator": "Exists", "effect": "NoSchedule"}},
				"securityContext":               map[string]any{"runAsUser": 0, "runAsGroup": 0, "seccompProfile": map[string]any{"type": "RuntimeDefault"}},
				"containers": []any{map[string]any{"name": "collector", "image": req.Image, "imagePullPolicy": "IfNotPresent", "env": env,
					"ports":           []any{map[string]any{"name": "management", "containerPort": 3421}},
					"securityContext": map[string]any{"allowPrivilegeEscalation": false, "readOnlyRootFilesystem": true, "capabilities": map[string]any{"drop": []string{"ALL"}}},
					"resources":       map[string]any{"requests": map[string]any{"cpu": "100m", "memory": "128Mi"}, "limits": map[string]any{"cpu": "1", "memory": "512Mi"}},
					"startupProbe":    probe("/healthz", 24), "livenessProbe": probe("/healthz", 3), "readinessProbe": probe("/readyz", 3),
					"volumeMounts": []any{map[string]any{"name": "pod-logs", "mountPath": "/var/log/pods", "readOnly": true},
						map[string]any{"name": "state", "mountPath": "/var/lib/ctrlb"}, map[string]any{"name": "token", "mountPath": "/var/run/secrets/ctrlb", "readOnly": true},
						map[string]any{"name": "tmp", "mountPath": "/tmp"}},
				}},
				"volumes": []any{map[string]any{"name": "pod-logs", "hostPath": map[string]any{"path": "/var/log/pods", "type": "Directory"}},
					map[string]any{"name": "state", "hostPath": map[string]any{"path": "/var/lib/ctrlb/installations/" + id, "type": "DirectoryOrCreate"}},
					map[string]any{"name": "token", "secret": map[string]any{"secretName": name, "defaultMode": 0400}}, map[string]any{"name": "tmp", "emptyDir": map[string]any{}}},
			}},
		}},
	}
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	for _, document := range documents {
		if err := encoder.Encode(document); err != nil {
			return nil, "", err
		}
	}
	if err := encoder.Close(); err != nil {
		return nil, "", err
	}
	return output.Bytes(), name + ".yaml", nil
}
