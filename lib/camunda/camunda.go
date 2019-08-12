package camunda

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"github.com/SmartEnergyPlatform/util/http/request"
	uuid "github.com/satori/go.uuid"
	"log"
	"time"
)

type Camunda struct {
	config util.ConfigType
	workerId string
}

func NewCamunda(config util.ConfigType) CamundaInterface {
	return &Camunda{config: config, workerId: uuid.NewV4().String()}
}


func (this *Camunda) GetTask() (tasks []messages.CamundaTask, err error) {
	fetchRequest := messages.CamundaFetchRequest{
		WorkerId: this.workerId,
		MaxTasks: this.config.CamundaWorkerTasks,
		Topics:   []messages.CamundaTopic{{LockDuration: this.config.CamundaFetchLockDuration, Name: this.config.CamundaTopic}},
	}
	err, _, _ = request.Post(this.config.CamundaUrl+"/external-task/fetchAndLock", fetchRequest, &tasks)
	return
}



func (this *Camunda) CompleteTask(taskId string, workerId string, outputName string, output messages.SenergyTask) (err error) {
	var completeRequest messages.CamundaCompleteRequest

	if workerId == "" {
		workerId = this.workerId
	}

	if this.config.CompletionStrategy == "pessimistic" {
		variables := map[string]messages.CamundaOutput{
			outputName: {
				Value: output,
			},
		}
		completeRequest = messages.CamundaCompleteRequest{WorkerId: workerId, Variables: variables}
	} else {
		completeRequest = messages.CamundaCompleteRequest{WorkerId: workerId}
		duration := time.Duration(this.config.OptimisticTaskCompletionTimeout) * time.Millisecond
		time.Sleep(duration)
	}


	pl := ""
	var code int
	log.Println("Start complete Request")
	err, pl, code = request.Post(this.config.CamundaUrl+"/external-task/"+taskId+"/complete", completeRequest, nil)
	if code == 204 || code == 200 {
		log.Println("complete camunda task: ", completeRequest, pl)
	} else {
		this.Error(messages.CamundaTask{Id: taskId}, pl)
		log.Println("Error on completeCamundaTask.")
	}
	return
}


func (this *Camunda) SetRetry(taskid string) {
	retry := messages.CamundaRetrySetRequest{Retries: 1}
	request.Put(this.config.CamundaUrl+"/external-task/"+taskid+"/retries", retry, nil)
}

func (this *Camunda) Error(task messages.CamundaTask, msg string) {
	errorMsg := messages.CamundaError{WorkerId: this.workerId, ErrorMessage: msg, Retries: 0, ErrorDetails: msg}
	log.Println("Send Error to Camunda: ", msg)
	log.Println(request.Post(this.config.CamundaUrl+"/external-task/"+task.Id+"/failure", errorMsg, nil))
	//this.completeCamundaTask(taskid, this.GetWorkerId(), "error", messages.SenergyTask{ErrorMsg:msg})
}

func (this *Camunda) GetWorkerId()(string) {
	return this.workerId
}
