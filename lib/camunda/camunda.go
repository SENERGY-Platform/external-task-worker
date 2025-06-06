/*
 * Copyright 2019 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package camunda

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda/interfaces"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda/shards"
	"github.com/SENERGY-Platform/external-task-worker/lib/com"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"github.com/SENERGY-Platform/process-incident-api/lib/client"
	"github.com/SENERGY-Platform/service-commons/pkg/cache"
	"github.com/SENERGY-Platform/service-commons/pkg/signal"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

type Camunda struct {
	config            util.Config
	workerId          string
	producer          com.ProducerInterface
	shards            Shards
	metrics           interfaces.Metrics
	topicmux          TopicMutex
	taskCompleteCache *cache.Cache
}

type Shards interface {
	GetShards() (result []string, err error)
	GetShardForUser(userId string) (shardUrl string, err error)
}

func NewCamundaWithShards(config util.Config, producer com.ProducerInterface, metrics interfaces.Metrics, shards Shards) (result *Camunda, err error) {
	c, err := cache.New(cache.Config{})
	if err != nil {
		return nil, err
	}
	return &Camunda{config: config, workerId: util.GetId(), producer: producer, shards: shards, metrics: metrics, taskCompleteCache: c}, nil
}

func NewCamunda(config util.Config, producer com.ProducerInterface, metrics interfaces.Metrics) (result *Camunda, err error) {
	c, err := cache.New(cache.Config{
		CacheInvalidationSignalHooks: map[cache.Signal]cache.ToKey{
			signal.Known.CacheInvalidationAll: nil,
		},
	})
	if err != nil {
		return result, err
	}
	s, err := shards.New(config.ShardsDb, c)
	if err != nil {
		return result, err
	}
	return NewCamundaWithShards(config, producer, metrics, s)
}

func (this *Camunda) ProvideTasks(ctx context.Context) (<-chan []messages.CamundaExternalTask, <-chan error, error) {
	shardList, err := this.shards.GetShards()
	if err != nil {
		this.metrics.LogCamundaGetShardsError()
		return nil, nil, err
	}
	tasks := make(chan []messages.CamundaExternalTask, 100)
	errChan := make(chan error, 100)
	wg := sync.WaitGroup{}
	for _, shard := range shardList {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					temp, err := this.getShardTasks(ctx, shard)
					if err != nil {
						this.metrics.LogCamundaGetTasksError()
						errChan <- err
						continue
					}
					this.metrics.LogCamundaLoadedTasks(len(temp))
					tasks <- temp
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(tasks)
		close(errChan)
	}()
	return tasks, errChan, nil
}

func (this *Camunda) getShardTasks(ctx context.Context, shard string) (tasks []messages.CamundaExternalTask, err error) {
	fetchRequest := messages.CamundaFetchRequest{
		WorkerId:             this.workerId,
		MaxTasks:             this.config.CamundaWorkerTasks,
		Topics:               []messages.CamundaTopic{{LockDuration: this.config.CamundaFetchLockDuration, Name: this.config.CamundaTopic}},
		AsyncResponseTimeout: this.config.CamundaLongPollTimeout,
	}
	if fetchRequest.AsyncResponseTimeout == 0 {
		fetchRequest.AsyncResponseTimeout = 10000 //10s
	}
	client := http.Client{Timeout: 15 * time.Second}
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(fetchRequest)
	if err != nil {
		return tasks, err
	}
	endpoint := shard + "/engine-rest/external-task/fetchAndLock"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, b)
	if err != nil {
		return tasks, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return tasks, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		temp, err := io.ReadAll(resp.Body)
		err = errors.New(fmt.Sprintln(endpoint, resp.Status, resp.StatusCode, string(temp), err))
		return tasks, err
	}
	err = json.NewDecoder(resp.Body).Decode(&tasks)
	return tasks, err
}

var UnableToCompleteErrResp = errors.New("unable to complete task")

func (this *Camunda) CompleteTask(taskInfo messages.TaskInfo, outputName string, output interface{}) (err error) {
	topic := taskInfo.TenantId + "+" + taskInfo.TaskId
	this.topicmux.Lock(topic)
	defer this.topicmux.Unlock(topic)
	//every process task complete may only happen once
	//use the cache.Use method to do the complete, only if the complete is not found in cache
	//cached by half the camunda lock duration to enable retries
	_, err = cache.Use[string](this.taskCompleteCache, topic, func() (string, error) {
		return "", this.completeTask(taskInfo, outputName, output)
	}, cache.NoValidation, time.Duration(this.config.CamundaFetchLockDuration/2)*time.Millisecond)
	return err
}

func (this *Camunda) completeTask(taskInfo messages.TaskInfo, outputName string, output interface{}) (err error) {
	shard, err := this.shards.GetShardForUser(taskInfo.TenantId)
	if err != nil {
		return err
	}

	var completeRequest messages.CamundaCompleteRequest

	if taskInfo.WorkerId == "" {
		taskInfo.WorkerId = this.workerId
	}
	if output != nil {
		variables := map[string]messages.CamundaOutput{
			outputName: {
				Value: output,
			},
		}
		completeRequest = messages.CamundaCompleteRequest{WorkerId: taskInfo.WorkerId, Variables: variables}
	} else {
		completeRequest = messages.CamundaCompleteRequest{WorkerId: taskInfo.WorkerId}
	}

	log.Println("Start complete Request", taskInfo.TaskId)
	send := func() error {
		client := http.Client{Timeout: 5 * time.Second}
		b := new(bytes.Buffer)
		err = json.NewEncoder(b).Encode(completeRequest)
		if err != nil {
			return err
		}
		resp, err := client.Post(shard+"/engine-rest/external-task/"+taskInfo.TaskId+"/complete", "application/json", b)
		if err != nil {
			this.metrics.LogCamundaCompleteTaskError()
			return err
		}
		defer resp.Body.Close()
		pl, err := io.ReadAll(resp.Body)
		if err != nil {
			this.metrics.LogCamundaCompleteTaskError()
			return err
		}
		if resp.StatusCode >= 300 {
			this.metrics.LogCamundaCompleteTaskError()
			log.Println("WARNING: unable to complete task", taskInfo.TaskId, taskInfo.ProcessInstanceId, taskInfo.ProcessDefinitionId, string(pl), taskInfo.TenantId)
			return fmt.Errorf("%w: %v, %v", UnableToCompleteErrResp, resp.StatusCode, string(pl))
		}
		this.metrics.LogCamundaCompleteTask()
		log.Println("complete camunda task: ", completeRequest, string(pl))
		return nil
	}

	err = send()
	if err != nil {
		log.Println("retry complete request")
		err = send()
		if err != nil && errors.Is(err, UnableToCompleteErrResp) {
			this.Error(taskInfo.TaskId, taskInfo.ProcessInstanceId, taskInfo.ProcessDefinitionId, err.Error(), taskInfo.TenantId)
		}
	}

	return
}

func (this *Camunda) UnlockTask(taskInfo messages.TaskInfo) (err error) {
	if this.config.Debug {
		log.Println("unlock task for retry", taskInfo.TaskId)
	}
	shard, err := this.shards.GetShardForUser(taskInfo.TenantId)
	if err != nil {
		return err
	}

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(shard+"/engine-rest/external-task/"+taskInfo.TaskId+"/unlock", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	pl, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		//this.Error(taskInfo.TaskId, taskInfo.ProcessInstanceId, taskInfo.ProcessDefinitionId, string(pl), taskInfo.TenantId)
		log.Println("WARNING: unable to unlock task", taskInfo.TaskId, taskInfo.ProcessInstanceId, taskInfo.ProcessDefinitionId, string(pl), taskInfo.TenantId)
	} else {
		log.Println("unlock camunda task: ", taskInfo)
	}
	return
}

func (this *Camunda) SetRetry(taskid string, tenantId string, retries int64) {
	shard, err := this.shards.GetShardForUser(tenantId)
	if err != nil {
		log.Println("ERROR: unable to get shard for SetRetry()", err)
		debug.PrintStack()
		return
	}
	retry := messages.CamundaRetrySetRequest{Retries: retries}

	client := http.Client{Timeout: 5 * time.Second}
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(retry)
	if err != nil {
		return
	}
	request, err := http.NewRequest("PUT", shard+"/engine-rest/external-task/"+taskid+"/retries", b)
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
	pl, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("ERROR: ReadAll():", err)
		debug.PrintStack()
		return
	}
	if resp.StatusCode >= 300 {
		log.Println("ERROR: unexpected SetRetry() response status", string(pl))
	}
}

func (this *Camunda) Error(externalTaskId string, processInstanceId string, processDefinitionId string, msg string, tenantId string) {
	this.metrics.LogIncident()
	incident := messages.Incident{
		Id:                  util.GetId(),
		ExternalTaskId:      externalTaskId,
		ProcessInstanceId:   processInstanceId,
		ProcessDefinitionId: processDefinitionId,
		WorkerId:            this.GetWorkerId(),
		ErrorMessage:        msg,
		Time:                util.TimeNow(),
		TenantId:            tenantId,
	}
	if this.config.UseHttpIncidentProducer && this.config.IncidentApiUrl != "" && this.config.IncidentApiUrl != "-" {
		err, _ := client.New(this.config.IncidentApiUrl).CreateIncident(client.InternalAdminToken, incident)
		if err != nil {
			log.Println("ERROR:", err)
			debug.PrintStack()
			return
		}
	} else {
		b, err := json.Marshal(messages.KafkaIncidentsCommand{
			Command:    "POST",
			MsgVersion: 3,
			Incident:   &incident,
		})
		if err != nil {
			log.Println("ERROR:", err)
			debug.PrintStack()
			return
		}
		err = this.producer.ProduceWithKey(this.config.KafkaIncidentTopic, processDefinitionId, string(b))
		if err != nil {
			log.Println("ERROR:", err)
			debug.PrintStack()
			return
		}
	}
}

func (this *Camunda) GetWorkerId() string {
	return this.workerId
}
