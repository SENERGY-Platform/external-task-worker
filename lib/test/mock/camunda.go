package mock

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda"
	"github.com/SENERGY-Platform/external-task-worker/lib/kafka"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"sync"
	"time"
)

var Camunda = &CamundaMock{}

type CamundaMock struct {
	waitingTasks   []messages.CamundaTask
	fetchedTasks   map[string]messages.CamundaTask
	completedTasks map[string]messages.Command
	failedTasks    map[string]messages.CamundaTask
	config         util.Config
	mux            sync.Mutex
	lockTimes      map[string]time.Time
}

func (this *CamundaMock) Get(config util.Config, producer kafka.ProducerInterface) camunda.CamundaInterface {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.waitingTasks = []messages.CamundaTask{}
	this.fetchedTasks = map[string]messages.CamundaTask{}
	this.completedTasks = map[string]messages.Command{}
	this.failedTasks = map[string]messages.CamundaTask{}
	this.lockTimes = map[string]time.Time{}
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

	//unlock old tasks
	for _, task := range this.fetchedTasks {
		timestamp, ok := this.lockTimes[task.Id]
		if ok && time.Now().Sub(timestamp) > time.Duration(this.config.CamundaFetchLockDuration)*time.Millisecond {
			this.waitingTasks = append(this.waitingTasks, task)
			delete(this.fetchedTasks, task.Id)
		}
	}

	size := int(this.config.CamundaWorkerTasks)
	if size > len(this.waitingTasks) {
		size = len(this.waitingTasks)
	}
	tasks = this.waitingTasks[:size]
	for _, task := range tasks {
		this.fetchedTasks[task.Id] = task
		this.lockTimes[task.Id] = time.Now()
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
	this.mux.Lock()
	defer this.mux.Unlock()
	temp, ok := this.fetchedTasks[taskid]
	if ok {
		temp.Retries = temp.Retries + 1
		this.fetchedTasks[taskid] = temp
	}
}

func (this *CamundaMock) Error(task messages.CamundaTask, msg string) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.failedTasks[task.Id] = task
	delete(this.fetchedTasks, task.Id)
}

func (this *CamundaMock) GetWorkerId() string {
	return "workerid"
}

func (this *CamundaMock) GetStatus() (fetched map[string]messages.CamundaTask, completed map[string]messages.Command, failed map[string]messages.CamundaTask) {
	this.mux.Lock()
	defer this.mux.Unlock()
	return this.fetchedTasks, this.completedTasks, this.failedTasks
}
