package operators

import (
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
)

type Operator interface {
	GetAgentInfo() (map[string]string, error)
	StartAgent() error
	StopAgent() error
	GracefulShutdown() error
	UpdateCurrentConfig(map[string]any) error
	GetMetrics() (map[string]any, error)
}

type OperatorService struct {
	Operator Operator
}

func NewOperatorService(adapter adapters.Adapter, agentType string) *OperatorService {
	var operator Operator
	if agentType == "fluentbit" {
		operator = NewFluentBitOperator(adapter)
	} else {
		operator = NewOtelOperator(adapter)
	}

	return &OperatorService{Operator: operator}
}

func (o *OperatorService) GetAgentInfo() (map[string]string, error) {
	return o.Operator.GetAgentInfo()
}

func (o *OperatorService) StartAgent() error {
	return o.Operator.StartAgent()
}

func (o *OperatorService) StopAgent() error {
	return o.Operator.StopAgent()
}

func (o *OperatorService) GracefulShutdown() error {
	return o.Operator.GracefulShutdown()
}

func (o *OperatorService) UpdateCurrentConfig(updateConfigRequest map[string]any) error {
	return o.Operator.UpdateCurrentConfig(updateConfigRequest)
}

func (o *OperatorService) GetMetrics() (map[string]any, error) {
	return o.Operator.GetMetrics()
}