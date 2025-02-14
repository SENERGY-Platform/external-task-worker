/*
 * Copyright 2020 InfAI (CC SES)
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

package test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda/interfaces"
	"github.com/SENERGY-Platform/external-task-worker/lib/com"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/mock"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestHealthCheckBy(t *testing.T) {
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.GroupScheduler = util.PARALLEL
	config.CompletionStrategy = util.PESSIMISTIC
	config.CamundaLongPollTimeout = 100

	freePort, err := getFreePort()
	if err != nil {
		t.Error(err)
		return
	}
	config.HealthCheckPort = strconv.Itoa(freePort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafka := &StoppableTestKafka{On: true}
	camunda := &StoppableTestCamunda{On: true}

	go lib.Worker(ctx, config, kafka, mock.Repo, camunda, mock.Marshaller, mock.Timescale)
	time.Sleep(1 * time.Second)

	mock.Repo.RegisterDevice(model.Device{
		Id:           "device_1",
		Name:         "d1",
		DeviceTypeId: "dt1",
		LocalId:      "d1u",
	})

	mock.Repo.RegisterProtocol(model.Protocol{
		Id:               "p1",
		Name:             "protocol1",
		Handler:          "protocol1",
		ProtocolSegments: []model.ProtocolSegment{{Id: "ms1", Name: "body"}},
	})

	mock.Repo.RegisterService(model.Service{
		Id:         "service_1",
		Name:       "on",
		LocalId:    "power",
		ProtocolId: "p1",
		Inputs: []model.Content{
			{
				Id: "metrics",
				ContentVariable: model.ContentVariable{
					Id:   "metrics",
					Name: "metrics",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:    "power",
							Name:  "power",
							Type:  model.Boolean,
							Value: true,
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	})

	cmd1 := messages.Command{
		Function:   model.Function{RdfType: model.SES_ONTOLOGY_CONTROLLING_FUNCTION},
		DeviceId:   "device_1",
		ServiceId:  "service_1",
		ProtocolId: "p1",
	}

	cmdMsg1, err := json.Marshal(cmd1)
	if err != nil {
		log.Fatal(err)
	}

	camunda.SetTask(messages.CamundaExternalTask{
		Id:       "1",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg1),
			},
		},
	})

	t.Log((time.Duration(config.CamundaLongPollTimeout) * time.Millisecond * 2).String())

	t.Run("check immediately", testHealtCheck(config, true))
	time.Sleep(time.Duration(config.CamundaLongPollTimeout) * time.Millisecond)
	t.Run("check after timeout", testHealtCheck(config, true))
	time.Sleep(10 * time.Duration(config.CamundaLongPollTimeout) * time.Millisecond)
	t.Run("check after 10 x timeout", testHealtCheck(config, true))

	kafka.On = false
	time.Sleep(4 * time.Duration(config.CamundaLongPollTimeout) * time.Millisecond)
	t.Run("check with kafka off", testHealtCheck(config, false))

	kafka.On = true
	time.Sleep(4 * time.Duration(config.CamundaLongPollTimeout) * time.Millisecond)
	t.Run("check with kafka reactivated", testHealtCheck(config, true))

	camunda.On = false
	time.Sleep(4 * time.Duration(config.CamundaLongPollTimeout) * time.Millisecond)
	t.Run("check with camunda off", testHealtCheck(config, false))

	camunda.On = true
	time.Sleep(4 * time.Duration(config.CamundaLongPollTimeout) * time.Millisecond)
	t.Run("check with camunda reactivated", testHealtCheck(config, true))

	camunda.On = false
	kafka.On = false
	time.Sleep(4 * time.Duration(config.CamundaLongPollTimeout) * time.Millisecond)
	t.Run("check with kafka and camunda off", testHealtCheck(config, false))

	camunda.On = true
	kafka.On = true
	time.Sleep(4 * time.Duration(config.CamundaLongPollTimeout) * time.Millisecond)
	t.Run("check with kafka and camunda reactivated", testHealtCheck(config, true))
}

func testHealtCheck(config util.Config, expectedOk bool) func(t *testing.T) {
	return func(t *testing.T) {
		resp, err := http.Get("http://localhost:" + config.HealthCheckPort)
		if err != nil {
			t.Error(err)
			return
		}
		var state lib.WorkerState
		err = json.NewDecoder(resp.Body).Decode(&state)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(state)
		if (resp.StatusCode == 200) != expectedOk {
			t.Error(expectedOk, resp.StatusCode)
		}
	}
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func GetFreePort() (string, error) {
	temp, err := getFreePort()
	return strconv.Itoa(temp), err
}

type StoppableTestKafka struct {
	On bool
}

func (this *StoppableTestKafka) Stop() {}

func (this *StoppableTestKafka) Produce(topic string, message string) (err error) {
	if this.On {
		return nil
	} else {
		return errors.New("producer is stopped")
	}
}

func (this *StoppableTestKafka) ProduceWithKey(topic string, message string, key string) (err error) {
	return this.Produce(topic, message)
}

func (this *StoppableTestKafka) Close() {}

func (this *StoppableTestKafka) NewConsumer(ctx context.Context, config util.Config, respListener func(msg string) error, errListener func(msg string) error) (err error) {
	return nil
}

func (this *StoppableTestKafka) NewProducer(ctx context.Context, config util.Config) (com.ProducerInterface, error) {
	return this, nil
}

func (this *StoppableTestKafka) Log(logger *log.Logger) {

}

type StoppableTestCamunda struct {
	On    bool
	tasks []messages.CamundaExternalTask
}

func (this *StoppableTestCamunda) Get(configType util.Config, producer com.ProducerInterface, metrics interfaces.Metrics) (interfaces.CamundaInterface, error) {
	return this, nil
}

func (this *StoppableTestCamunda) ProvideTasks(ctx context.Context) (<-chan []messages.CamundaExternalTask, <-chan error, error) {
	tasks := make(chan []messages.CamundaExternalTask, 100)
	errChan := make(chan error, 100)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Millisecond * 100)
				temp, err := this.GetTasks()
				if err != nil {
					errChan <- err
					continue
				}
				tasks <- temp
			}
		}
	}()
	go func() {
		wg.Wait()
		close(tasks)
		close(errChan)
	}()
	return tasks, errChan, nil
}

func (this *StoppableTestCamunda) GetTasks() (tasks []messages.CamundaExternalTask, err error) {
	if this.On {
		return this.tasks, nil
	} else {
		return []messages.CamundaExternalTask{}, errors.New("producer is stopped")
	}
}

func (this *StoppableTestCamunda) CompleteTask(taskInfo messages.TaskInfo, outputName string, output interface{}) (err error) {
	return nil
}

func (this *StoppableTestCamunda) SetRetry(taskid string, tenantId string, number int64) {}

func (this *StoppableTestCamunda) Error(externalTaskId string, processInstanceId string, processDefinitionId string, msg string, tenantId string) {
}

func (this *StoppableTestCamunda) GetWorkerId() string {
	return ""
}

func (this *StoppableTestCamunda) SetTask(task messages.CamundaExternalTask) {
	this.tasks = []messages.CamundaExternalTask{task}
}

func (this *StoppableTestCamunda) UnlockTask(taskInfo messages.TaskInfo) (err error) {
	return nil
}
