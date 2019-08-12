package camunda

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
)

type FactoryInterface interface {
	Get(configType util.ConfigType)CamundaInterface
}

type CamundaInterface interface{
	GetTask() (tasks []messages.CamundaTask, err error)
	CompleteTask(taskId string, workerId string, outputName string, output messages.Command) (err error)
	SetRetry(taskid string)
	Error(task messages.CamundaTask, msg string)
	GetWorkerId() string
}


