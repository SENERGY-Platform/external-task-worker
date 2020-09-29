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
	"fmt"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/mock"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"strconv"
	"time"
)

func ExampleIncidents() {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	idCount := 0
	util.GetId = func() string {
		idCount = idCount + 1
		return strconv.Itoa(idCount)
	}

	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	kafka, err := mock.Kafka.NewProducer(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	camunda := camunda.NewCamundaWithShards(config, kafka, nil)
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
