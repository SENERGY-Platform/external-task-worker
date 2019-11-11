package mock

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda"
	"github.com/SENERGY-Platform/external-task-worker/lib/kafka"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"sync"
	"time"
)

var Camunda = &CamundaMock{}

type CamundaMock struct {
	waitingTasks   []messages.CamundaExternalTask
	fetchedTasks   map[string]messages.CamundaExternalTask
	completedTasks map[string]messages.Command
	failedTasks    map[string]messages.CamundaExternalTask
	config         util.Config
	mux            sync.Mutex
	lockTimes      map[string]time.Time
}

func (this *CamundaMock) Get(config util.Config, producer kafka.ProducerInterface) camunda.CamundaInterface {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.waitingTasks = []messages.CamundaExternalTask{}
	this.fetchedTasks = map[string]messages.CamundaExternalTask{}
	this.completedTasks = map[string]messages.Command{}
	this.failedTasks = map[string]messages.CamundaExternalTask{}
	this.lockTimes = map[string]time.Time{}
	this.config = config
	return this
}

func (this *CamundaMock) AddTask(task messages.CamundaExternalTask) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.waitingTasks = append(this.waitingTasks, task)
}

func (this *CamundaMock) GetTask() (tasks []messages.CamundaExternalTask, err error) {
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

func (this *CamundaMock) SetRetry(taskid string, number int64) {
	log.Println("DEBUG: SetRetry", taskid, number)
	this.mux.Lock()
	defer this.mux.Unlock()
	temp, ok := this.fetchedTasks[taskid]
	if ok {
		temp.Retries = number
		this.fetchedTasks[taskid] = temp
	}
}

func (this *CamundaMock) Error(taskId string, msg string) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.failedTasks[taskId] = this.fetchedTasks[taskId]
	delete(this.fetchedTasks, taskId)
}

func (this *CamundaMock) GetWorkerId() string {
	return "workerid"
}

func (this *CamundaMock) GetStatus() (fetched map[string]messages.CamundaExternalTask, completed map[string]messages.Command, failed map[string]messages.CamundaExternalTask) {
	this.mux.Lock()
	defer this.mux.Unlock()
	return this.fetchedTasks, this.completedTasks, this.failedTasks
}
