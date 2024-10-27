package operators

import (
	"github.com/ctrlb-hq/ctrlb-collector/internal/adapters"
	"github.com/ctrlb-hq/ctrlb-collector/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/internal/models"
	"github.com/ctrlb-hq/ctrlb-collector/internal/utils"
)

type Operator interface {
	UpdateCurrentConfig(models.UpdateConfigRequest) (map[string]string, error)
	StartAgent() (map[string]string, error)
	StopAgent() (map[string]string, error)
	GracefulShutdown() (map[string]string, error)
	CurrentStatus() (*models.AgentMetrics, error)
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

func (o *OperatorService) UpdateCurrentConfig(updateConfigRequest models.UpdateConfigRequest) (interface{}, error) {
	return o.Operator.UpdateCurrentConfig(updateConfigRequest)
}

func (o *OperatorService) GetCurrentConfig() (interface{}, error) {
	return utils.LoadYAML(constants.AGENT_CONFIG_PATH)
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

func (o *OperatorService) CurrentStatus() (*models.AgentMetrics, error) {
	return o.Operator.CurrentStatus()
}
