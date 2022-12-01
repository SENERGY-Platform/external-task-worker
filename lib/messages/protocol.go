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

package messages

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"time"
)

type TaskInfo struct {
	WorkerId            string `json:"worker_id"`
	TaskId              string `json:"task_id"`
	ProcessInstanceId   string `json:"process_instance_id"`
	ProcessDefinitionId string `json:"process_definition_id"`
	CompletionStrategy  string `json:"completion_strategy"`
	Time                string `json:"time"`
	TenantId            string `json:"tenant_id"`
}

type ProtocolRequest struct {
	Input map[string]string `json:"input"`
}

type ProtocolResponse struct {
	Output map[string]string `json:"output"`
}

type Metadata struct {
	Version              int64             `json:"version,omitempty"`
	Device               model.Device      `json:"device"`
	Service              model.Service     `json:"service"`
	Protocol             model.Protocol    `json:"protocol"`
	OutputPath           string            `json:"output_path,omitempty"`        //only for version >= 3
	OutputFunctionId     string            `json:"output_function_id,omitempty"` //only for version >= 3 if no OutputPath is known
	OutputAspectNode     *model.AspectNode `json:"output_aspect_node,omitempty"` //only for version >= 3 if no OutputPath is known
	InputCharacteristic  string            `json:"input_characteristic,omitempty"`
	OutputCharacteristic string            `json:"output_characteristic,omitempty"`
	ContentVariableHints []string          `json:"content_variable_hints,omitempty"` //only for version < 3
	ResponseTo           string            `json:"response_to"`
	ErrorTo              string            `json:"error_to,omitempty"`
}

type ProtocolMsg struct {
	Request  ProtocolRequest  `json:"request"`
	Response ProtocolResponse `json:"response"`
	TaskInfo TaskInfo         `json:"task_info"`
	Metadata Metadata         `json:"metadata"`
	Trace    []Trace          `json:"trace,omitempty"`
}

type Trace struct {
	TimeUnit  string `json:"time_unit"`
	Timestamp int64  `json:"timestamp"`
	Location  string `json:"location"`
}

type KafkaMessage struct {
	Topic   string
	Key     string
	Payload string
}

type EventRequest struct {
	Device           model.Device
	Service          model.Service
	Protocol         model.Protocol
	CharacteristicId string
	FunctionId       string
	AspectNode       model.AspectNode
}

type RequestInfo struct {
	Request      *KafkaMessage
	Event        *EventRequest
	Metadata     GroupTaskMetadataElement
	SubTaskState SubTaskState
}

type RequestInfoList []RequestInfo

type SubTaskState struct {
	LastTry  time.Time
	TryCount int64
}

type TimescaleRequest struct {
	Device     model.Device
	Service    model.Service
	ColumnName string
}

type TimescaleResponse struct {
	Time  *string     `json:"time"`
	Value interface{} `json:"value"`
}
