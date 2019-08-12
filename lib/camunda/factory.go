package camunda

import "github.com/SENERGY-Platform/external-task-worker/util"

type FactoryType struct {}

var Factory FactoryType

func (FactoryType) Get(config util.ConfigType) CamundaInterface {
	return NewCamunda(config)
}
