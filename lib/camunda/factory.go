package camunda

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/kafka"
	"github.com/SENERGY-Platform/external-task-worker/util"
)

type FactoryType struct{}

var Factory FactoryType

func (FactoryType) Get(config util.Config, producer kafka.ProducerInterface) CamundaInterface {
	return NewCamunda(config, producer)
}
