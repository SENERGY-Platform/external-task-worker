package mock

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"sync"
)

var Camunda = &CamundaMock{}

type CamundaMock struct {
	waitingTasks   []messages.CamundaTask
	fetchedTasks   map[string]messages.CamundaTask
	completedTasks map[string]messages.Command
	failedTasks    map[string]messages.CamundaTask
	config         util.Config
	mux            sync.Mutex
}

func (this *CamundaMock) Get(config util.Config) camunda.CamundaInterface {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.waitingTasks = []messages.CamundaTask{}
	this.fetchedTasks = map[string]messages.CamundaTask{}
	this.completedTasks = map[string]messages.Command{}
	this.failedTasks = map[string]messages.CamundaTask{}
	this.config = config
	return this
}

func (this *CamundaMock) AddTask(task messages.CamundaTask) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.waitingTasks = append(this.waitingTasks, task)
}

func (this *CamundaMock) GetTask() (tasks []messages.CamundaTask, err error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	size := int(this.config.CamundaWorkerTasks)
	if size > len(this.waitingTasks) {
		size = len(this.waitingTasks)
	}
	tasks = this.waitingTasks[:size]
	for _, task := range tasks {
		this.fetchedTasks[task.Id] = task
	}
	this.waitingTasks = this.waitingTasks[size:]
	return tasks, nil
}

func (this *CamundaMock) CompleteTask(taskId string, workerId string, outputName string, output messages.Command) (err error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	_, ok := this.fetchedTasks[taskId]
	if !ok {
		return errors.New("task not found " + taskId)
	}
	delete(this.fetchedTasks, taskId)
	this.completedTasks[taskId] = output
	return
}

func (this *CamundaMock) SetRetry(taskid string) {

}

func (this *CamundaMock) Error(task messages.CamundaTask, msg string) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.failedTasks[task.Id] = task
}

func (this *CamundaMock) GetWorkerId() string {
	return "workerid"
}

func (this *CamundaMock) GetStatus() (map[string]messages.CamundaTask, map[string]messages.Command, map[string]messages.CamundaTask) {
	this.mux.Lock()
	defer this.mux.Unlock()
	return this.fetchedTasks, this.completedTasks, this.failedTasks
}
