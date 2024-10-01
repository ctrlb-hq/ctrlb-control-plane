package services

import (
	"errors"

	"github.com/ctrlb-hq/ctrlb-collector/internal/adapters"
	"github.com/ctrlb-hq/ctrlb-collector/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/internal/utils"
)

type OperatorService struct {
	FluentBitOperator FluentBitOperator
}

func NewOperatorService(adapter adapters.Adapter) *OperatorService {
	fluentBitOperator := NewFluentBitOperator(adapter)
	return &OperatorService{FluentBitOperator: *fluentBitOperator}
}

func (o *OperatorService) GetUptime() (map[string]interface{}, error) {
	switch constants.AGENT_TYPE {
	case "fluent-bit":
		return o.FluentBitOperator.GetUptime()
	default:
		return nil, errors.New("agent type not supported yet")
	}
}

func (o *OperatorService) UpdateCurrentConfig(updateConfigRequest interface{}) (interface{}, error) {
	switch constants.AGENT_TYPE {
	case "fluent-bit":
		return o.FluentBitOperator.UpdateCurrentConfig(updateConfigRequest)
	default:
		return nil, errors.New("agent type not supported yet")
	}
}

func (o *OperatorService) GetCurrentConfig() (interface{}, error) {
	return utils.LoadYAMLToJSON(constants.AGENT_CONFIG_PATH)
}

func (o *OperatorService) StartAgent() (map[string]string, error) {
	switch constants.AGENT_TYPE {
	case "fluent-bit":
		return o.FluentBitOperator.StartAgent()
	default:
		return nil, errors.New("agent type not supported yet")
	}
}

func (o *OperatorService) StopAgent() (map[string]string, error) {
	switch constants.AGENT_TYPE {
	case "fluent-bit":
		return o.FluentBitOperator.StopAgent()
	default:
		return nil, errors.New("agent type not supported yet")
	}
}

func (o *OperatorService) GracefulShutdown() (map[string]string, error) {
	switch constants.AGENT_TYPE {
	case "fluent-bit":
		return o.FluentBitOperator.GracefulShutdown()
	default:
		return nil, errors.New("agent type not supported yet")
	}
}

func (o *OperatorService) CurrentStatus() (map[string]string, error) {
	switch constants.AGENT_TYPE {
	case "fluent-bit":
		return o.FluentBitOperator.CurrentStatus()
	default:
		return nil, errors.New("agent type not supported yet")
	}
}
