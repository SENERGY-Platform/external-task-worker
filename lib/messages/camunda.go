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
	"github.com/SENERGY-Platform/process-incident-api/lib/client"
)

type CamundaVariable struct {
	Type  string      `json:"type,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type CamundaOutput struct {
	Type  string      `json:"type,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

// https://github.com/camunda/camunda-docs-manual/blob/master/content/reference/rest/external-task/fetch.md
type CamundaExternalTask struct {
	Id                  string                     `json:"id,omitempty"`
	Variables           map[string]CamundaVariable `json:"variables,omitempty"`
	ActivityId          string                     `json:"activityId,omitempty"`
	Retries             int64                      `json:"retries"`
	ExecutionId         string                     `json:"executionId"`
	ProcessInstanceId   string                     `json:"processInstanceId"`
	ProcessDefinitionId string                     `json:"processDefinitionId"`
	TenantId            string                     `json:"tenantId"`
	Error               string                     `json:"errorMessage"`
}

type CamundaTopic struct {
	Name         string `json:"topicName,omitempty"`
	LockDuration int64  `json:"lockDuration,omitempty"`
}

type CamundaFetchRequest struct {
	WorkerId             string         `json:"workerId,omitempty"`
	MaxTasks             int64          `json:"maxTasks,omitempty"`
	Topics               []CamundaTopic `json:"topics,omitempty"`
	AsyncResponseTimeout int64          `json:"asyncResponseTimeout,omitempty"`
}

// https://github.com/camunda/camunda-docs-manual/blob/master/content/reference/rest/external-task/post-complete.md
type CamundaCompleteRequest struct {
	WorkerId  string                   `json:"workerId,omitempty"`
	Variables map[string]CamundaOutput `json:"localVariables,omitempty"`
}

type CamundaRetrySetRequest struct {
	Retries int64 `json:"retries"`
}

type KafkaIncidentsCommand struct {
	Command             string    `json:"command"`
	MsgVersion          int64     `json:"msg_version"`
	Incident            *Incident `json:"incident,omitempty"`
	ProcessDefinitionId string    `json:"process_definition_id,omitempty"`
	ProcessInstanceId   string    `json:"process_instance_id,omitempty"`
}

type Incident = client.IncidentMessage

type TaskWithDuplicateIndex struct {
	Task           CamundaExternalTask
	DuplicateIndex int
}
