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

package test

import (
	"context"
	"fmt"
	"github.com/SENERGY-Platform/external-task-worker/lib"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/lib/prometheus"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/mock"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestUserIncident(t *testing.T) {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	idCount := 0
	temp := util.GetId
	util.GetId = func() string {
		idCount = idCount + 1
		return strconv.Itoa(idCount)
	}
	defer func() {
		util.GetId = temp
	}()
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.CompletionStrategy = util.OPTIMISTIC

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := camunda.NewCamundaWithShards(config, mock.Kafka, prometheus.NewMetrics("test", nil), nil)
	if err != nil {
		log.Fatal(err)
	}
	mockCamunda := &mock.CamundaMock{Camunda: c}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller, mock.Timescale)

	time.Sleep(1 * time.Second)

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id: "1",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_INCIDENT: {
				Value: "test error",
			},
		},
		ActivityId:          "aid1",
		Retries:             0,
		ExecutionId:         "execid1",
		ProcessInstanceId:   "piid1",
		ProcessDefinitionId: "pdid1",
		TenantId:            "user",
	})

	time.Sleep(1 * time.Second)

	incidents := mock.Kafka.GetProduced(config.KafkaIncidentTopic)

	expected := []string{`{"command":"POST","msg_version":3,"incident":{"id":"2","external_task_id":"1","process_instance_id":"piid1","process_definition_id":"pdid1","worker_id":"1","error_message":"user triggered incident: test error","time":"0001-01-01T00:00:00Z","tenant_id":"user","deployment_name":""}}`}

	if !reflect.DeepEqual(incidents, expected) {
		t.Errorf("expected != actual\n%#v\n%#v\n", expected, incidents)
	}

}

func Example_camunda_Error() {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	idCount := 0
	temp := util.GetId
	util.GetId = func() string {
		idCount = idCount + 1
		return strconv.Itoa(idCount)
	}
	defer func() {
		util.GetId = temp
	}()

	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafka, err := mock.Kafka.NewProducer(ctx, config)
	if err != nil {
		fmt.Println(err)
		return
	}

	camunda, err := camunda.NewCamundaWithShards(config, kafka, prometheus.NewMetrics("test", nil), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	camunda.Error("task_id_1", "piid_1", "pdid_1", "error message", "user1")
	camunda.Error("task_id_2", "piid_2", "pdid_2", "error message", "user1")

	incidents := mock.Kafka.GetProduced(config.KafkaIncidentTopic)

	for _, message := range incidents {
		fmt.Println(message)
	}

	//output:
	//{"command":"POST","msg_version":3,"incident":{"id":"2","external_task_id":"task_id_1","process_instance_id":"piid_1","process_definition_id":"pdid_1","worker_id":"1","error_message":"error message","time":"0001-01-01T00:00:00Z","tenant_id":"user1","deployment_name":""}}
	//{"command":"POST","msg_version":3,"incident":{"id":"3","external_task_id":"task_id_2","process_instance_id":"piid_2","process_definition_id":"pdid_2","worker_id":"1","error_message":"error message","time":"0001-01-01T00:00:00Z","tenant_id":"user1","deployment_name":""}}
}

func Example_camunda_ErrorOverHttp() {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	idCount := 0
	temp := util.GetId
	util.GetId = func() string {
		idCount = idCount + 1
		return strconv.Itoa(idCount)
	}
	defer func() {
		util.GetId = temp
	}()

	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafka, err := mock.Kafka.NewProducer(ctx, config)
	if err != nil {
		fmt.Println(err)
		return
	}

	incidentsApiMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg, _ := io.ReadAll(r.Body)
		fmt.Printf("http incident: %v %v %v\n", r.Method, r.URL.Path, string(msg))
	}))
	defer incidentsApiMockServer.Close()
	config.IncidentApiUrl = incidentsApiMockServer.URL
	config.UseHttpIncidentProducer = true

	camunda, err := camunda.NewCamundaWithShards(config, kafka, prometheus.NewMetrics("test", nil), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	camunda.Error("task_id_1", "piid_1", "pdid_1", "error message", "user1")
	camunda.Error("task_id_2", "piid_2", "pdid_2", "error message", "user1")

	incidents := mock.Kafka.GetProduced(config.KafkaIncidentTopic)

	for _, message := range incidents {
		fmt.Println("kafka incident:", message)
	}

	//output:
	//http incident: POST /incidents {"id":"2","external_task_id":"task_id_1","process_instance_id":"piid_1","process_definition_id":"pdid_1","worker_id":"1","error_message":"error message","time":"0001-01-01T00:00:00Z","tenant_id":"user1","deployment_name":""}
	//http incident: POST /incidents {"id":"3","external_task_id":"task_id_2","process_instance_id":"piid_2","process_definition_id":"pdid_2","worker_id":"1","error_message":"error message","time":"0001-01-01T00:00:00Z","tenant_id":"user1","deployment_name":""}
}
