package configcompiler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
)

// CompileGraphToJSON preserves directed processing order. Each signal follows a
// linear processor chain, with receiver fan-in and exporter fan-out supported.
func CompileGraphToJSON(graph models.PipelineGraph) (*map[string]any, error) {
	receivers, processors, exporters, pipelines, err := buildPipelines(graph)
	if err != nil {
		return nil, err
	}
	service := map[string]any{"pipelines": pipelines, "telemetry": constants.TelemetryService}
	config := map[string]any{"receivers": receivers, "processors": processors, "exporters": exporters, "service": service}
	extensions := map[string]any{}
	for _, node := range graph.Nodes {
		if node.ComponentRole != "receiver" || componentType(node.ComponentName, node.ComponentRole) != "file_log" {
			continue
		}
		storage, exists := node.Config["storage"]
		if !exists {
			continue
		}
		name, ok := storage.(string)
		if ok && name == "" {
			continue
		}
		if !ok || (name != "file_storage" && !strings.HasPrefix(name, "file_storage/")) {
			return nil, fmt.Errorf("file log storage must reference file_storage or a named file_storage instance")
		}
		extensions[name] = map[string]any{"directory": "${env:STATE_DIR}/offsets", "create_directory": true}
	}
	if len(extensions) > 0 {
		names := make([]string, 0, len(extensions))
		for name := range extensions {
			names = append(names, name)
		}
		sort.Strings(names)
		config["extensions"], service["extensions"] = extensions, names
	}
	return &config, nil
}

// Schema names carry a final role suffix. Internal underscores belong to the
// component type; old schema spellings remain compatible with existing graphs.
func componentType(name, role string) string {
	name = strings.TrimSuffix(name, "_"+role)
	aliases := map[string]string{
		"filelog": "file_log", "hostmetrics": "host_metrics", "kubeletstats": "kubelet_stats",
		"k8sattributes": "k8s_attributes", "memorylimiter": "memory_limiter",
		"probabilisticsampler": "probabilistic_sampler", "tailsampling": "tail_sampling",
		"otlp_grpc": "otlp", "otlphttp": "otlp_http",
	}
	if canonical, ok := aliases[name]; ok {
		return canonical
	}
	return name
}

func nodeAlias(node models.PipelineNodes) (string, error) {
	typ := componentType(node.ComponentName, node.ComponentRole)
	if typ == "" || strings.ContainsAny(typ, "/ \t\r\n") {
		return "", fmt.Errorf("invalid component type")
	}
	data, err := json.Marshal(node.Config)
	if err != nil {
		return "", fmt.Errorf("component configuration is not JSON serializable")
	}
	hash := sha256.Sum256(data)
	name := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.' {
			return r
		}
		return '_'
	}, strings.TrimSpace(node.Name))
	if name == "" {
		name = "component"
	}
	return fmt.Sprintf("%s/%s_%s", typ, name, hex.EncodeToString(hash[:6])), nil
}

func supports(node models.PipelineNodes, signal string) bool {
	for _, value := range node.SupportedSignals {
		if value == signal {
			return true
		}
	}
	return false
}

type graphNode struct {
	node  models.PipelineNodes
	alias string
	next  []string
}

type signalPath struct {
	signal     string
	receivers  []string
	processors []string
	exporters  []string
}

func buildPipelines(graph models.PipelineGraph) (map[string]any, map[string]any, map[string]any, Pipelines, error) {
	if len(graph.Nodes) == 0 {
		return nil, nil, nil, nil, fmt.Errorf("empty pipeline graph")
	}
	nodes := map[string]*graphNode{}
	aliases := map[string]bool{}
	ids := make([]string, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		id := strconv.Itoa(node.ComponentID)
		if node.ComponentID <= 0 || nodes[id] != nil {
			return nil, nil, nil, nil, fmt.Errorf("component IDs must be unique positive integers")
		}
		if node.ComponentRole != "receiver" && node.ComponentRole != "processor" && node.ComponentRole != "exporter" {
			return nil, nil, nil, nil, fmt.Errorf("unknown component role")
		}
		if len(node.SupportedSignals) == 0 {
			return nil, nil, nil, nil, fmt.Errorf("each component must support at least one signal")
		}
		for _, signal := range node.SupportedSignals {
			if signal != "logs" && signal != "metrics" && signal != "traces" {
				return nil, nil, nil, nil, fmt.Errorf("unknown telemetry signal")
			}
		}
		alias, err := nodeAlias(node)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		if aliases[alias] {
			return nil, nil, nil, nil, fmt.Errorf("component names must distinguish nodes with identical types and configurations")
		}
		aliases[alias] = true
		nodes[id] = &graphNode{node: node, alias: alias}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return nodes[ids[i]].alias < nodes[ids[j]].alias })
	edges := map[string]bool{}
	for _, edge := range graph.Edges {
		from, to := nodes[edge.Source], nodes[edge.Target]
		if from == nil || to == nil {
			return nil, nil, nil, nil, fmt.Errorf("edge references a nonexistent component")
		}
		if from.node.ComponentRole == "exporter" || to.node.ComponentRole == "receiver" {
			return nil, nil, nil, nil, fmt.Errorf("edges must lead from receivers through processors to exporters")
		}
		key := edge.Source + ":" + edge.Target
		if edges[key] {
			return nil, nil, nil, nil, fmt.Errorf("duplicate graph edge")
		}
		edges[key] = true
		from.next = append(from.next, edge.Target)
	}
	state := map[string]int{}
	var visit func(string) error
	visit = func(id string) error {
		if state[id] == 1 {
			return fmt.Errorf("pipeline graph contains a cycle")
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		node := nodes[id]
		sort.Slice(node.next, func(i, j int) bool { return nodes[node.next[i]].alias < nodes[node.next[j]].alias })
		for _, next := range node.next {
			if err := visit(next); err != nil {
				return err
			}
		}
		state[id] = 2
		return nil
	}
	for _, id := range ids {
		if err := visit(id); err != nil {
			return nil, nil, nil, nil, err
		}
	}
	paths := map[string]*signalPath{}
	used := map[string]bool{}
	for _, id := range ids {
		if nodes[id].node.ComponentRole != "receiver" {
			continue
		}
		for _, signal := range []string{"logs", "metrics", "traces"} {
			if !supports(nodes[id].node, signal) {
				continue
			}
			current := id
			chain := []string{}
			exporters := []string{}
			pathIDs := []string{id}
			for {
				processors, destinations := []string{}, []string{}
				for _, next := range nodes[current].next {
					if !supports(nodes[next].node, signal) {
						continue
					}
					if nodes[next].node.ComponentRole == "processor" {
						processors = append(processors, next)
					} else {
						destinations = append(destinations, next)
					}
				}
				if len(processors) > 1 || (len(processors) > 0 && len(destinations) > 0) {
					return nil, nil, nil, nil, fmt.Errorf("ambiguous processor branch for %s: use one processor chain before exporter fan-out", signal)
				}
				if len(processors) == 1 {
					current = processors[0]
					chain = append(chain, nodes[current].alias)
					pathIDs = append(pathIDs, current)
					continue
				}
				for _, destination := range destinations {
					exporters = append(exporters, nodes[destination].alias)
					pathIDs = append(pathIDs, destination)
				}
				break
			}
			if len(exporters) == 0 {
				continue
			}
			for _, pathID := range pathIDs {
				used[pathID] = true
			}
			keyJSON, _ := json.Marshal([]any{signal, chain, exporters})
			key := string(keyJSON)
			if paths[key] == nil {
				paths[key] = &signalPath{signal: signal, processors: chain, exporters: exporters}
			}
			paths[key].receivers = append(paths[key].receivers, nodes[id].alias)
		}
	}
	receivers, processors, exporters := map[string]any{}, map[string]any{}, map[string]any{}
	for _, id := range ids {
		if !used[id] {
			return nil, nil, nil, nil, fmt.Errorf("every component must belong to a supported receiver-to-exporter path")
		}
		node := nodes[id]
		switch node.node.ComponentRole {
		case "receiver":
			receivers[node.alias] = node.node.Config
		case "processor":
			processors[node.alias] = node.node.Config
		case "exporter":
			exporters[node.alias] = node.node.Config
		}
	}
	pipelines := Pipelines{}
	for _, path := range paths {
		sort.Strings(path.receivers)
		pipeline := Pipeline{Receivers: path.receivers, Processors: path.processors, Exporters: path.exporters}
		data, _ := json.Marshal(pipeline)
		hash := sha256.Sum256(data)
		name := path.signal + "/pipeline_" + hex.EncodeToString(hash[:6])
		pipelines[name] = pipeline
	}
	return receivers, processors, exporters, pipelines, nil
}
