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

package interfaces

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/kafka"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
)

type FactoryInterface interface {
	Get(configType util.Config, producer kafka.ProducerInterface) (CamundaInterface, error)
}

type CamundaInterface interface {
	GetTasks() (tasks []messages.CamundaExternalTask, err error)
	CompleteTask(taskInfo messages.TaskInfo, outputName string, output interface{}) (err error)
	SetRetry(taskid string, tenantId string, number int64)
	Error(externalTaskId string, processInstanceId string, processDefinitionId string, msg string, tenantId string)
	GetWorkerId() string
}
