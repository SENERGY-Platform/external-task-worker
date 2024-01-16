package camunda

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda/shards"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/lib/prometheus"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/docker"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/mock"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGetTask(t *testing.T) {
	temp := util.GetId
	util.GetId = func() string {
		return "test-worker"
	}
	defer func() {
		util.GetId = temp
	}()

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	defer wg.Wait()
	defer cancel()

	pgConn, err := docker.Postgres(ctx, &wg, "test")
	if err != nil {
		t.Error(err)
		return
	}

	s, err := shards.New(pgConn, nil)
	if err != nil {
		t.Error(err)
		return
	}

	_, camundaPgIp, _, err := docker.PostgresWithNetwork(ctx, &wg, "camunda")
	if err != nil {
		t.Error(err)
		return
	}

	camundaUrl, err := docker.Camunda(ctx, &wg, camundaPgIp, "5432")
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

	_, err = mock.Kafka.NewProducer(ctx, util.Config{})
	if err != nil {
		t.Fatal(err)
	}

	var pid string
	var tasks []messages.CamundaExternalTask
	t.Run("create normal process", testCreateProcess(camundaUrl, &pid))
	t.Run("start process", testStartDeployment(shard, pid))
	t.Run("run get task", testGetTasks(s, &tasks))
	t.Run("complete task", testCompleteTask(s, tasks))
	t.Run("check incidents", testCheckIncidents())
}

func testCheckIncidents() func(t *testing.T) {
	return func(t *testing.T) {
		if incidents := mock.Kafka.GetProduced("incidents"); len(incidents) > 0 {
			t.Fatal(incidents)
		}
	}
}

func testCompleteTask(s *shards.Shards, tasks []messages.CamundaExternalTask) func(t *testing.T) {
	return func(t *testing.T) {
		if len(tasks) < 1 {
			t.Fatal(tasks)
		}
		err := NewCamundaWithShards(util.Config{
			KafkaIncidentTopic: "incidents",
		}, mock.Kafka, prometheus.NewMetrics("test", nil), s).CompleteTask(messages.TaskInfo{
			WorkerId:            "test-worker",
			TaskId:              tasks[0].Id,
			ProcessInstanceId:   tasks[0].ProcessInstanceId,
			ProcessDefinitionId: tasks[0].ProcessDefinitionId,
			TenantId:            tasks[0].TenantId,
		}, "result", map[string]interface{}{"r": 255, "g": 0, "b": 100})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testGetTasks(s *shards.Shards, tasks *[]messages.CamundaExternalTask) func(t *testing.T) {
	return func(t *testing.T) {
		var err error
		*tasks, err = NewCamundaWithShards(util.Config{
			CamundaFetchLockDuration: 60000,
			CamundaTopic:             "pessimistic",
			CamundaWorkerTasks:       10,
		}, mock.Kafka, prometheus.NewMetrics("test", nil), s).GetTasks()
		if err != nil {
			t.Fatal(err)
		}
		if len(*tasks) != 1 {
			t.Fatal(*tasks)
		}
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
	err = json.NewDecoder(resp.Body).Decode(&result)
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
		"id": "urn:infai:ses:measuring-function:bdb6a7c8-4a3d-4fe0-bab3-ce02e09b5869",
		"name": "",
		"concept_id": "",
		"rdf_type": ""
	},
	"characteristic_id": "urn:infai:ses:characteristic:5b4eea52-e8e5-4e80-9455-0382f81a1b43",
	"aspect": {
		"id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6",
		"name": "",
		"rdf_type": ""
	},
	"device_id": "hue",
	"service_id": "urn:infai:ses:service:99614933-4734-41b6-a131-3f96f134ee69",
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
