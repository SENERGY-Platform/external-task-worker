package camunda

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/kafka"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"
)

type Camunda struct {
	config   util.Config
	workerId string
	producer kafka.ProducerInterface
}

func NewCamunda(config util.Config, producer kafka.ProducerInterface) CamundaInterface {
	return &Camunda{config: config, workerId: uuid.NewV4().String(), producer: producer}
}

func (this *Camunda) GetTask() (tasks []messages.CamundaTask, err error) {
	fetchRequest := messages.CamundaFetchRequest{
		WorkerId: this.workerId,
		MaxTasks: this.config.CamundaWorkerTasks,
		Topics:   []messages.CamundaTopic{{LockDuration: this.config.CamundaFetchLockDuration, Name: this.config.CamundaTopic}},
	}
	client := http.Client{Timeout: 5 * time.Second}
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(fetchRequest)
	if err != nil {
		return
	}
	resp, err := client.Post(this.config.CamundaUrl+"/external-task/fetchAndLock", "application/json", b)
	if err != nil {
		return tasks, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&tasks)
	return
}

func (this *Camunda) CompleteTask(taskId string, workerId string, outputName string, output messages.Command) (err error) {
	var completeRequest messages.CamundaCompleteRequest

	if workerId == "" {
		workerId = this.workerId
	}

	if this.config.CompletionStrategy == util.PESSIMISTIC {
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

	log.Println("Start complete Request")
	client := http.Client{Timeout: 5 * time.Second}
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(completeRequest)
	if err != nil {
		return
	}
	resp, err := client.Post(this.config.CamundaUrl+"/external-task/"+taskId+"/complete", "application/json", b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	pl, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		this.Error(messages.CamundaTask{Id: taskId}, string(pl))
		log.Println("Error on completeCamundaTask.")
	} else {
		log.Println("complete camunda task: ", completeRequest, string(pl))
	}
	return
}

func (this *Camunda) SetRetry(taskid string, retries int64) {
	retry := messages.CamundaRetrySetRequest{Retries: retries}

	client := http.Client{Timeout: 5 * time.Second}
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(retry)
	if err != nil {
		return
	}
	request, err := http.NewRequest("PUT", this.config.CamundaUrl+"/external-task/"+taskid+"/retries", b)
	if err != nil {
		log.Println("ERROR: SetRetry():", err)
		debug.PrintStack()
		return
	}
	request.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(request)
	if err != nil {
		log.Println("ERROR: SetRetry():", err)
		debug.PrintStack()
		return
	}
	defer resp.Body.Close()
	pl, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ERROR: ReadAll():", err)
		debug.PrintStack()
		return
	}
	if resp.StatusCode >= 300 {
		log.Println("ERROR: unexpected SetRetry() response status", string(pl))
	}
}

func (this *Camunda) Error(task messages.CamundaTask, msg string) {
	var err error
	if task.ProcessInstanceId == "" {
		task, err = this.readTaskById(task.Id)
		if err != nil {
			return
		}
	}
	_ = this.sendErrorToCamunda(msg, task)

	_ = this.sendErrorToKafka(msg, task)

	_ = this.removeProcessInstance(task.ProcessInstanceId)

	return
}

func (this *Camunda) sendErrorToCamunda(msg string, task messages.CamundaTask) error {
	errorMsg := messages.CamundaError{WorkerId: this.workerId, ErrorMessage: msg, Retries: 0, ErrorDetails: msg}
	if this.config.Debug {
		log.Println("Send Error to Camunda: ", msg)
	}
	client := http.Client{Timeout: 5 * time.Second}
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(errorMsg)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	resp, err := client.Post(this.config.CamundaUrl+"/external-task/"+task.Id+"/failure", "application/json", b)
	if err != nil {
		log.Println("ERROR: camunda.Error():", err)
		debug.PrintStack()
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		pl, _ := ioutil.ReadAll(resp.Body)
		err = errors.New("ERROR: unexpected camunda.Error() response status: " + resp.Status + " " + string(pl))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	return err
}

func (this *Camunda) GetWorkerId() string {
	return this.workerId
}

func (this *Camunda) sendErrorToKafka(msg string, task messages.CamundaTask) error {
	b, err := json.Marshal(messages.KafkaIncidentMessage{WorkerId: this.workerId, ErrorMessage: msg, Retries: 0, Task: task})
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	this.producer.Produce(this.config.KafkaIncidentTopic, string(b))
	return nil
}

func (this *Camunda) removeProcessInstance(id string) (err error) {
	client := &http.Client{Timeout: 5 * time.Second}
	request, err := http.NewRequest("DELETE", this.config.CamundaUrl+"/engine-rest/process-instance/"+url.PathEscape(id)+"?skipIoMappings=true", nil)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	resp, err := client.Do(request)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	defer resp.Body.Close()
	if !(resp.StatusCode == 200 || resp.StatusCode == 204) {
		msg, _ := ioutil.ReadAll(resp.Body)
		err = errors.New("error on delete in engine for " + this.config.CamundaUrl + "/engine-rest/process-instance/" + url.PathEscape(id) + ": " + resp.Status + " " + string(msg))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	return nil
}

func (this *Camunda) readTaskById(taskId string) (task messages.CamundaTask, err error) {
	client := &http.Client{Timeout: 5 * time.Second}
	request, err := http.NewRequest("GET", this.config.CamundaUrl+"/engine-rest/task/"+url.PathEscape(taskId), nil)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return task, err
	}
	resp, err := client.Do(request)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return task, err
	}
	defer resp.Body.Close()
	if !(resp.StatusCode == 200 || resp.StatusCode == 204) {
		msg, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(resp.Status + " " + string(msg))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return task, err
	}
	err = json.NewDecoder(resp.Body).Decode(&task)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return task, err
	}
	return task, err
}
