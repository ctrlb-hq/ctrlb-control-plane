package services

import (
	"github.com/ctrlb-hq/ctrlb-collector/internal/adapters"
	"github.com/ctrlb-hq/ctrlb-collector/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/internal/utils"
)

type Operator interface {
	GetUptime() (map[string]interface{}, error)
	UpdateCurrentConfig(interface{}) (map[string]string, error)
	StartAgent() (map[string]string, error)
	StopAgent() (map[string]string, error)
	GracefulShutdown() (map[string]string, error)
	CurrentStatus() (map[string]string, error)
}

type OperatorService struct {
	Operator Operator
}

func NewOperatorService(adapter adapters.Adapter) *OperatorService {
	var operator Operator

	switch constants.AGENT_TYPE {
	case "fluent-bit":
		operator = NewFluentBitOperator(adapter)
	case "otel":
		operator = NewOtelOperator(adapter)
	default:
		return nil
	}
	return &OperatorService{Operator: operator}

}

func (o *OperatorService) GetUptime() (map[string]interface{}, error) {
	return o.Operator.GetUptime()
}

func (o *OperatorService) UpdateCurrentConfig(updateConfigRequest interface{}) (interface{}, error) {
	return o.Operator.UpdateCurrentConfig(updateConfigRequest)
}

func (o *OperatorService) GetCurrentConfig() (interface{}, error) {
	return utils.LoadYAMLToJSON(constants.AGENT_CONFIG_PATH)
}

func (o *OperatorService) StartAgent() (map[string]string, error) {
	return o.Operator.StartAgent()
}

func (o *OperatorService) StopAgent() (map[string]string, error) {
	return o.Operator.StopAgent()
}

func (o *OperatorService) GracefulShutdown() (map[string]string, error) {
	return o.Operator.GracefulShutdown()
}

func (o *OperatorService) CurrentStatus() (map[string]string, error) {
	return o.Operator.CurrentStatus()
}
