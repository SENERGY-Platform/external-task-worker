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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda/shards"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/docker"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/mock"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestWorkerErrorRetries(t *testing.T) {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.GroupScheduler = util.PARALLEL
	config.CompletionStrategy = util.PESSIMISTIC
	config.CamundaWorkerTimeout = 1000
	config.CamundaFetchLockDuration = 60000
	config.HttpCommandConsumerPort = ""
	config.Debug = true
	config.CamundaTopic = "pessimistic"
	mock.CleanKafkaMock()

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, camundaPgIp, _, err := docker.PostgresWithNetwork(ctx, wg, "camunda")
	if err != nil {
		t.Error(err)
		return
	}

	camundaUrl, err := docker.Camunda(ctx, wg, camundaPgIp, "5432")
	if err != nil {
		t.Error(err)
		return
	}

	pgConn, err := docker.Postgres(ctx, wg, "test")
	if err != nil {
		t.Error(err)
		return
	}

	s, err := shards.New(pgConn, nil)
	if err != nil {
		t.Error(err)
		return
	}

	err = s.EnsureShard(camundaUrl)
	if err != nil {
		t.Error(err)
		return
	}
	shard, err := s.EnsureShardForUser("owner")
	if err != nil {
		t.Error(err)
		return
	}

	config.ShardsDb = pgConn

	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, camunda.Factory, mock.Marshaller, mock.Timescale)

	time.Sleep(1 * time.Second)

	protocolMessages := []messages.ProtocolMsg{}
	mock.Kafka.Subscribe("protocol1", func(msg string) error {
		msgObj := messages.ProtocolMsg{}
		err := json.Unmarshal([]byte(msg), &msgObj)
		if err != nil {
			t.Error(err)
			return nil
		}
		protocolMessages = append(protocolMessages, msgObj)
		if msgObj.Response.Output == nil {
			msgObj.Response.Output = map[string]string{}
		}
		msgObj.Response.Output["error"] = "test error message"
		errResp, err := json.Marshal(msgObj)
		if err != nil {
			t.Error(err)
			return nil
		}
		go mock.Kafka.Produce(msgObj.Metadata.ErrorTo, string(errResp))
		return nil
	})

	//populate repository
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
		Name:       "s1",
		LocalId:    "s1u",
		ProtocolId: "p1",
		Outputs: []model.Content{
			{
				Id: "metrics",
				ContentVariable: model.ContentVariable{
					Id:   "metrics",
					Name: "metrics",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:               "level",
							Name:             "level",
							Type:             model.Integer,
							CharacteristicId: example.Hex,
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	})

	var pid string
	t.Run("create normal process", testCreateProcess(camundaUrl, &pid))
	t.Run("start process", testStartDeployment(shard, pid))

	time.Sleep(10 * time.Second)
	temp, _ := json.Marshal(protocolMessages)
	t.Log(len(protocolMessages), string(temp))

	if len(protocolMessages) != 3 {
		t.Error(len(protocolMessages), string(temp))
	}
}

func testCreateProcess(url string, pidRef *string) func(t *testing.T) {
	return func(t *testing.T) {
		result, err := deployProcess(url, "test", bpmnExample, svgExample, "owner", "test")
		if err != nil {
			t.Fatal(err)
		}
		deploymentId, ok := result["id"].(string)
		if !ok {
			t.Fatal(err)
		}
		*pidRef = deploymentId
	}
}

func deployProcess(shard string, name string, xml string, svg string, owner string, source string) (result map[string]interface{}, err error) {
	result = map[string]interface{}{}
	boundary := "---------------------------" + time.Now().String()
	b := strings.NewReader(buildPayLoad(name, xml, svg, boundary, owner, source))
	resp, err := http.Post(shard+"/engine-rest/deployment/create", "multipart/form-data; boundary="+boundary, b)
	if err != nil {
		log.Println("ERROR: request to processengine ", err)
		return result, err
	}
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		return nil, errors.New(resp.Status + " " + string(temp))
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	log.Println("deployed process:", result)
	return
}

func testStartDeployment(shard string, id string) func(t *testing.T) {
	return func(t *testing.T) {
		definitions, err := getRawDefinitionsByDeployment(shard, id)
		if err != nil {
			t.Fatal(err)
		}
		if len(definitions) == 0 {
			t.Fatal("no definition for deployment found")
		}
		err = startProcessGetId(shard, definitions[0].Id)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func startProcessGetId(shard string, processDefinitionId string) (err error) {
	body, err := json.Marshal(map[string]string{})
	if err != nil {
		return err
	}
	resp, err := http.Post(shard+"/engine-rest/process-definition/"+url.QueryEscape(processDefinitionId)+"/start", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}
	temp, _ := io.ReadAll(resp.Body)
	log.Println("start process:", resp.StatusCode, string(temp))
	return
}

type ProcessDefinition struct {
	Id                string `json:"id,omitempty"`
	Key               string `json:"key,omitempty"`
	Category          string `json:"category,omitempty"`
	Description       string `json:"description,omitempty"`
	Name              string `json:"name,omitempty"`
	Version           int    `json:"Version,omitempty"`
	Resource          string `json:"resource,omitempty"`
	DeploymentId      string `json:"deploymentId,omitempty"`
	Diagram           string `json:"diagram,omitempty"`
	Suspended         bool   `json:"suspended,omitempty"`
	TenantId          string `json:"tenantId,omitempty"`
	VersionTag        string `json:"versionTag,omitempty"`
	HistoryTimeToLive int    `json:"historyTimeToLive,omitempty"`
}

func getRawDefinitionsByDeployment(shard string, deploymentId string) (result []ProcessDefinition, err error) {
	resp, err := http.Get(shard + "/engine-rest/process-definition?deploymentId=" + url.QueryEscape(deploymentId))
	if err != nil {
		return result, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

func buildPayLoad(name string, xml string, svg string, boundary string, owner string, deploymentSource string) string {
	segments := []string{}
	if deploymentSource == "" {
		deploymentSource = "sepl"
	}

	segments = append(segments, "Content-Disposition: form-data; name=\"data\"; "+"filename=\""+name+".bpmn\"\r\nContent-Type: text/xml\r\n\r\n"+xml+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"diagram\"; "+"filename=\""+name+".svg\"\r\nContent-Type: image/svg+xml\r\n\r\n"+svg+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"deployment-name\"\r\n\r\n"+name+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"deployment-source\"\r\n\r\n"+deploymentSource+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"tenant-id\"\r\n\r\n"+owner+"\r\n")

	return "--" + boundary + "\r\n" + strings.Join(segments, "--"+boundary+"\r\n") + "--" + boundary + "--\r\n"
}

const bpmnExample = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI" xmlns:dc="http://www.omg.org/spec/DD/20100524/DC" xmlns:camunda="http://camunda.org/schema/1.0/bpmn" xmlns:senergy="https://senergy.infai.org" xmlns:di="http://www.omg.org/spec/DD/20100524/DI" id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
    <bpmn:process id="ExampleId" name="ExampleName" isExecutable="true" senergy:description="ExampleDesc">
        <bpmn:startEvent id="StartEvent_1">
            <bpmn:outgoing>SequenceFlow_0qjn3dq</bpmn:outgoing>
        </bpmn:startEvent>
        <bpmn:sequenceFlow id="SequenceFlow_0qjn3dq" sourceRef="StartEvent_1" targetRef="Task_1nbnl8y"/>
        <bpmn:sequenceFlow id="SequenceFlow_15v8030" sourceRef="Task_1nbnl8y" targetRef="Task_1lhzy95"/>
        <bpmn:endEvent id="EndEvent_1vhaxdr">
            <bpmn:incoming>SequenceFlow_17lypcn</bpmn:incoming>
        </bpmn:endEvent>
        <bpmn:sequenceFlow id="SequenceFlow_17lypcn" sourceRef="Task_1lhzy95" targetRef="EndEvent_1vhaxdr"/>
        <bpmn:serviceTask id="Task_1nbnl8y" name="Lighting getColorFunction" camunda:type="external" camunda:topic="pessimistic">
            <bpmn:extensionElements>
                <camunda:inputOutput>
                    <camunda:inputParameter name="payload"><![CDATA[{
	"function": {
		"id": "somefunction",
		"name": "",
		"concept_id": "",
		"rdf_type": ""
	},
	"characteristic_id": "example_hex",
	"aspect": {
		"id": "someaspect",
		"name": "",
		"rdf_type": ""
	},
	"device_id": "device_1",
	"service_id": "service_1",
	"input": {},
	"retries": 2
}]]></camunda:inputParameter>
                    <camunda:outputParameter name="outputs.b">${result.b}</camunda:outputParameter>
                    <camunda:outputParameter name="outputs.g">${result.g}</camunda:outputParameter>
                    <camunda:outputParameter name="outputs.r">${result.r}</camunda:outputParameter>
                </camunda:inputOutput>
            </bpmn:extensionElements>
            <bpmn:incoming>SequenceFlow_0qjn3dq</bpmn:incoming>
            <bpmn:outgoing>SequenceFlow_15v8030</bpmn:outgoing>
        </bpmn:serviceTask>
        <bpmn:serviceTask id="Task_1lhzy95" name="Lamp setColorFunction" camunda:type="external" camunda:topic="optimistic">
            <bpmn:extensionElements>
                <camunda:inputOutput>
                    <camunda:inputParameter name="payload"><![CDATA[{
	"function": {
		"id": "urn:infai:ses:controlling-function:c54e2a89-1fb8-4ecb-8993-a7b40b355599",
		"name": "",
		"concept_id": "",
		"rdf_type": ""
	},
	"characteristic_id": "urn:infai:ses:characteristic:5b4eea52-e8e5-4e80-9455-0382f81a1b43",
	"device_class": {
		"id": "urn:infai:ses:device-class:14e56881-16f9-4120-bb41-270a43070c86",
		"name": "",
		"image": "",
		"rdf_type": ""
	},
	"device_id": "hue",
	"service_id": "urn:infai:ses:service:67789396-d1ca-4ea9-9147-0614c6d68a2f",
	"input": {
		"b": 0,
		"g": 0,
		"r": 0
	}
}]]></camunda:inputParameter>
                    <camunda:inputParameter name="inputs.b">0</camunda:inputParameter>
                    <camunda:inputParameter name="inputs.g">255</camunda:inputParameter>
                    <camunda:inputParameter name="inputs.r">100</camunda:inputParameter>
                </camunda:inputOutput>
            </bpmn:extensionElements>
            <bpmn:incoming>SequenceFlow_15v8030</bpmn:incoming>
            <bpmn:outgoing>SequenceFlow_17lypcn</bpmn:outgoing>
        </bpmn:serviceTask>
    </bpmn:process>
    <bpmndi:BPMNDiagram id="BPMNDiagram_1">
        <bpmndi:BPMNPlane id="BPMNPlane_1" bpmnElement="ExampleId">
            <bpmndi:BPMNShape id="_BPMNShape_StartEvent_2" bpmnElement="StartEvent_1">
                <dc:Bounds x="173" y="102" width="36" height="36"/>
            </bpmndi:BPMNShape>
            <bpmndi:BPMNEdge id="SequenceFlow_0qjn3dq_di" bpmnElement="SequenceFlow_0qjn3dq">
                <di:waypoint x="209" y="120"/>
                <di:waypoint x="260" y="120"/>
            </bpmndi:BPMNEdge>
            <bpmndi:BPMNEdge id="SequenceFlow_15v8030_di" bpmnElement="SequenceFlow_15v8030">
                <di:waypoint x="360" y="120"/>
                <di:waypoint x="420" y="120"/>
            </bpmndi:BPMNEdge>
            <bpmndi:BPMNShape id="EndEvent_1vhaxdr_di" bpmnElement="EndEvent_1vhaxdr">
                <dc:Bounds x="582" y="102" width="36" height="36"/>
            </bpmndi:BPMNShape>
            <bpmndi:BPMNEdge id="SequenceFlow_17lypcn_di" bpmnElement="SequenceFlow_17lypcn">
                <di:waypoint x="520" y="120"/>
                <di:waypoint x="582" y="120"/>
            </bpmndi:BPMNEdge>
            <bpmndi:BPMNShape id="ServiceTask_04qer1y_di" bpmnElement="Task_1nbnl8y">
                <dc:Bounds x="260" y="80" width="100" height="80"/>
            </bpmndi:BPMNShape>
            <bpmndi:BPMNShape id="ServiceTask_0o8bz5b_di" bpmnElement="Task_1lhzy95">
                <dc:Bounds x="420" y="80" width="100" height="80"/>
            </bpmndi:BPMNShape>
        </bpmndi:BPMNPlane>
    </bpmndi:BPMNDiagram>
</bpmn:definitions>`
const svgExample = `<svg height='48' version='1.1' viewBox='167 96 134 48' width='134' xmlns='http://www.w3.org/2000/svg' xmlns:xlink='http://www.w3.org/1999/xlink'><defs><marker id='sequenceflow-end-white-black-3gh21e50i1p8scvqmvrotmp9p' markerHeight='10' markerWidth='10' orient='auto' refX='11' refY='10' viewBox='0 0 20 20'><path d='M 1 5 L 11 10 L 1 15 Z' style='fill: black; stroke-width: 1px; stroke-linecap: round; stroke-dasharray: 10000, 1; stroke: black;'/></marker></defs><g class='djs-group'><g class='djs-element djs-connection' data-element-id='SequenceFlow_04zz9eb' style='display: block;'><g class='djs-visual'><path d='m  209,120L259,120 ' style='fill: none; stroke-width: 2px; stroke: black; stroke-linejoin: round; marker-end: url(&apos;#sequenceflow-end-white-black-3gh21e50i1p8scvqmvrotmp9p&apos;);'/></g><polyline class='djs-hit' points='209,120 259,120 ' style='fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;'/><rect class='djs-outline' height='12' style='fill: none;' width='62' x='203' y='114'/></g></g><g class='djs-group'><g class='djs-element djs-shape' data-element-id='StartEvent_1' style='display: block;' transform='translate(173 102)'><g class='djs-visual'><circle cx='18' cy='18' r='18' style='stroke: black; stroke-width: 2px; fill: white; fill-opacity: 0.95;'/><path d='m 8.459999999999999,11.34 l 0,12.6 l 18.900000000000002,0 l 0,-12.6 z l 9.450000000000001,5.4 l 9.450000000000001,-5.4' style='fill: white; stroke-width: 1px; stroke: black;'/></g><rect class='djs-hit' height='36' style='fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;' width='36' x='0' y='0'></rect><rect class='djs-outline' height='48' style='fill: none;' width='48' x='-6' y='-6'></rect></g></g><g class='djs-group'><g class='djs-element djs-shape' data-element-id='EndEvent_056p30q' style='display: block;' transform='translate(259 102)'><g class='djs-visual'><circle cx='18' cy='18' r='18' style='stroke: black; stroke-width: 4px; fill: white; fill-opacity: 0.95;'/></g><rect class='djs-hit' height='36' style='fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;' width='36' x='0' y='0'></rect><rect class='djs-outline' height='48' style='fill: none;' width='48' x='-6' y='-6'></rect></g></g></svg>`
