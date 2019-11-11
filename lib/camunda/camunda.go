package camunda

import (
	"bytes"
	"encoding/json"
	"github.com/SENERGY-Platform/external-task-worker/lib/kafka"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"log"
	"net/http"
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

func (this *Camunda) GetTask() (tasks []messages.CamundaExternalTask, err error) {
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
		this.Error(taskId, string(pl))
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

func (this *Camunda) Error(externalTaskId string, msg string) {
	b, err := json.Marshal(messages.KafkaIncidentMessage{
		MsgVersion:     1,
		ExternalTaskId: externalTaskId,
		WorkerId:       this.GetWorkerId(),
		ErrorMessage:   msg,
		Time:           time.Now(),
	})
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return
	}
	this.producer.Produce(this.config.KafkaIncidentTopic, string(b))
}

func (this *Camunda) GetWorkerId() string {
	return this.workerId
}
