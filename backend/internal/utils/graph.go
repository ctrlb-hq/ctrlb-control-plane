package utils

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Pipeline Pipeline `yaml:"pipeline"`
}

type Pipeline struct {
	Inputs  []map[string]interface{} `yaml:"inputs"`
	Filters []map[string]interface{} `yaml:"filters"`
	Outputs []map[string]interface{} `yaml:"outputs"`
}

func parseConfig(configStr string) (*Config, error) {
	var config Config
	err := yaml.Unmarshal([]byte(configStr), &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func getUniqueKey(m map[string]map[string]interface{}, key string) string {
	if _, exists := m[key]; !exists {
		return key
	}

	for i := 1; ; i++ {
		newKey := fmt.Sprintf("%s_%d", key, i)
		if _, exists := m[newKey]; !exists {
			return newKey
		}
	}
}

func addInput(input map[string]interface{}, inputRecords map[string]map[string]interface{}) {
	inputName := getUniqueKey(inputRecords, input["name"].(string))
	input["type"] = "input"
	inputRecords[inputName] = input
}

func addFlbFilter(filter map[string]interface{}, edges *[]map[string]string, inputRecords, filterRecords map[string]map[string]interface{}) {
	filterName := getUniqueKey(filterRecords, filter["name"].(string))
	filter["type"] = "filter"
	filterRecords[filterName] = filter

	match := filter["match"].(string)
	if match == "*" {
		for _, input := range inputRecords {
			*edges = append(*edges, map[string]string{"src": input["name"].(string), "des": filterName})
			filters, ok := input["filters"].([]string)
			if !ok {
				filters = []string{}
			}
			input["filters"] = append(filters, filterName)
			input["rewrite_block"] = !filter["rewrite_keep"].(bool)
		}
	} else if strings.Contains(match, "*") {
		pattern := regexp.MustCompile("^" + regexp.QuoteMeta(match[:len(match)-2]) + "\\.")
		for _, input := range inputRecords {
			if pattern.MatchString(input["tag"].(string)) {
				*edges = append(*edges, map[string]string{"src": input["name"].(string), "des": filterName})
				filters, ok := input["filters"].([]string)
				if !ok {
					filters = []string{}
				}
				input["filters"] = append(filters, filterName)
				input["rewrite_block"] = !filter["rewrite_keep"].(bool)
			}
		}
	} else {
		var matchFound bool
		for _, input := range inputRecords {
			if input["tag"].(string) == match {
				matchFound = true
				*edges = append(*edges, map[string]string{"src": input["name"].(string), "des": filterName})
				filters, ok := input["filters"].([]string)
				if !ok {
					filters = []string{}
				}
				input["filters"] = append(filters, filterName)
				input["rewrite_block"] = !filter["rewrite_keep"].(bool)
				break
			}
		}
		if !matchFound {
			log.Fatal("NameError: No matching input found for filter")
		}
	}

	if rewriteKeep, ok := filter["rewrite_keep"].(bool); ok && rewriteKeep {
		*edges = append(*edges, map[string]string{"src": filterName, "des": filter["rewrite_keep_name"].(string)})
	}
}

func addFlbOutput(output map[string]interface{}, edges *[]map[string]string, inputRecords, outputRecords map[string]map[string]interface{}) {
	outputName := getUniqueKey(outputRecords, output["name"].(string))
	output["type"] = "output"
	outputRecords[outputName] = output

	match := output["match"].(string)
	if match == "*" {
		for _, input := range inputRecords {
			if rewriteBlock, ok := input["rewrite_block"].(bool); !ok || !rewriteBlock {
				if filters, ok := input["filters"].([]string); ok && len(filters) > 0 {
					for _, filter := range filters {
						*edges = append(*edges, map[string]string{"src": filter, "des": outputName})
					}
				} else {
					*edges = append(*edges, map[string]string{"src": input["name"].(string), "des": outputName})
				}
			}
		}
	} else if strings.Contains(match, "*") {
		pattern := regexp.MustCompile("^" + regexp.QuoteMeta(match[:len(match)-2]) + "\\.")
		for _, input := range inputRecords {
			if pattern.MatchString(input["tag"].(string)) {
				if rewriteBlock, ok := input["rewrite_block"].(bool); !ok || !rewriteBlock {
					if filters, ok := input["filters"].([]string); ok && len(filters) > 0 {
						for _, filter := range filters {
							*edges = append(*edges, map[string]string{"src": filter, "des": outputName})
						}
					} else {
						*edges = append(*edges, map[string]string{"src": input["name"].(string), "des": outputName})
					}
				}
			}
		}
	} else {
		var matchFound bool
		for _, input := range inputRecords {
			if input["tag"].(string) == match {
				if rewriteBlock, ok := input["rewrite_block"].(bool); !ok || !rewriteBlock {
					matchFound = true
					if filters, ok := input["filters"].([]string); ok && len(filters) > 0 {
						for _, filter := range filters {
							*edges = append(*edges, map[string]string{"src": filter, "des": outputName})
						}
					} else {
						*edges = append(*edges, map[string]string{"src": input["name"].(string), "des": outputName})
					}
					break
				}
			}
		}
		if !matchFound {
			log.Fatal("NameError: No matching input found for output")
		}
	}
}

func createFlbGraph(config *Config) map[string]interface{} {
	inputs := make(map[string]map[string]interface{})
	outputs := make(map[string]map[string]interface{})
	filters := make(map[string]map[string]interface{})
	var edges []map[string]string

	for _, input := range config.Pipeline.Inputs {
		addInput(input, inputs)
	}

	for i := range config.Pipeline.Filters {
		if config.Pipeline.Filters[i]["name"].(string) == "rewrite_tag" {
			rules := strings.Split(config.Pipeline.Filters[i]["rule"].(string), " ")
			emitterName, ok := config.Pipeline.Filters[i]["emitter_name"].(string)
			if !ok || emitterName == "" {
				emitterName = getUniqueKey(inputs, "rewrite_tag")
			}
			inputName := getUniqueKey(inputs, emitterName)
			inputs[inputName] = map[string]interface{}{"name": inputName, "tag": rules[2], "type": "input"}
			config.Pipeline.Filters[i]["rewrite_keep"] = rules[3] == "true"
			config.Pipeline.Filters[i]["rewrite_keep_name"] = inputName
		}
	}

	for _, filter := range config.Pipeline.Filters {
		addFlbFilter(filter, &edges, inputs, filters)
	}

	for _, output := range config.Pipeline.Outputs {
		addFlbOutput(output, &edges, inputs, outputs)
	}

	var nodes []string
	for key := range inputs {
		nodes = append(nodes, key)
	}
	for key := range filters {
		nodes = append(nodes, key)
	}
	for key := range outputs {
		nodes = append(nodes, key)
	}

	return map[string]interface{}{
		"nodes": nodes,
		"edges": edges,
	}
}

func DrawGraph(configStr string, agentType string) (map[string]interface{}, error) {
	config, err := parseConfig(configStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %v", err)
	}

	if agentType == "fluent-bit" {
		return createFlbGraph(config), nil
	} else {
		return nil, fmt.Errorf("error while creating graph: agent is not supported yet")
	}
}
