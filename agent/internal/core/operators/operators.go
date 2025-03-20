package operators

import (
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
)

type Operator interface {
	StartAgent() error
	StopAgent() error
	GracefulShutdown() error
	UpdateCurrentConfig(map[string]any) error
}

type OperatorService struct {
	Operator Operator
}

func NewOperatorService(adapter adapters.Adapter) *OperatorService {
	operator := NewOtelOperator(adapter)

	return &OperatorService{Operator: operator}
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
