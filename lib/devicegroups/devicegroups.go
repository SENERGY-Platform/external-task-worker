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

package devicegroups

import (
	"encoding/json"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/bradfitz/gomemcache/memcache"
	"log"
	"strconv"
	"strings"
)

const METADATA_KEY_PREFIX = "meta."
const RESULT_KEY_PREFIX = "result."

type Callback = func(command messages.Command, task messages.CamundaExternalTask) (topic string, key string, message string, err error)
type DbInterface interface {
	Get(key string) (*memcache.Item, error)
	Set(item *memcache.Item) error
}

func New(devicerepo devicerepository.RepoInterface, protocolMessageCallback Callback, expirationInSeconds int32, memcachedUrls []string) *DeviceGroups {
	if len(memcachedUrls) == 0 {
		log.Println("WARNING: start with local sub result storage")
		return NewWithKeyValueStore(devicerepo, protocolMessageCallback, expirationInSeconds, NewLocalDb())
	} else {
		return NewWithKeyValueStore(devicerepo, protocolMessageCallback, expirationInSeconds, memcache.New(memcachedUrls...))
	}
}

func NewWithKeyValueStore(devicerepo devicerepository.RepoInterface, protocolMessageCallback Callback, expirationInSeconds int32, db DbInterface) *DeviceGroups {
	return &DeviceGroups{
		protocolMessageCallback: protocolMessageCallback,
		repo:                    devicerepo,
		db:                      db,
		expirationInSeconds:     expirationInSeconds,
	}
}

type DeviceGroups struct {
	protocolMessageCallback Callback
	repo                    devicerepository.RepoInterface
	db                      DbInterface
	expirationInSeconds     int32
}

func (this *DeviceGroups) ProcessResponse(subTaskId string, subResult interface{}) (parent messages.GroupTaskMetadataElement, results []interface{}, finished bool, err error) {
	err = this.setSubResult(subTaskId, subResult)
	if err != nil {
		return parent, nil, false, err
	}
	return this.IsFinished(subTaskId)
}

func (this *DeviceGroups) ProcessCommand(request messages.Command, task messages.CamundaExternalTask) (missingRequests []messages.KafkaMessage, finishedResults []interface{}, err error) {
	var missingSubTasks []messages.GroupTaskMetadataElement
	_, finishedResults, missingSubTasks, err = this.getTaskResults(task.Id)
	if err == ErrNotFount {
		err = nil
		missingSubTasks, err = this.GetSubTasks(request, task)
		if err != nil {
			return nil, nil, err
		}
		for _, subTask := range missingSubTasks {
			err = this.setGroupMetadata(subTask.Task.Id, messages.GroupTaskMetadata{
				Parent: messages.GroupTaskMetadataElement{
					Command: request,
					Task:    task,
				},
				Children: missingSubTasks,
			})
			if err != nil {
				return nil, nil, err
			}
		}
	}
	for _, subTask := range missingSubTasks {
		protocolTopic, key, message, err := this.protocolMessageCallback(subTask.Command, subTask.Task)
		if err != nil {
			return nil, nil, err
		}
		missingRequests = append(missingRequests, messages.KafkaMessage{
			Topic:   protocolTopic,
			Key:     key,
			Payload: message,
		})
	}
	return
}

func (this *DeviceGroups) IsFinished(taskId string) (parent messages.GroupTaskMetadataElement, results []interface{}, finished bool, err error) {
	meta, results, missing, err := this.getTaskResults(taskId)
	if err == ErrNotFount {
		err = nil
		return parent, results, false, err
	}
	if err != nil {
		return parent, nil, false, err
	}
	finished = len(missing) > 0
	return meta.Parent, results, finished, err
}

func (this *DeviceGroups) getGroupMetadata(taskId string) (metadata messages.GroupTaskMetadata, err error) {
	err = this.dbGet(METADATA_KEY_PREFIX+taskId, &metadata)
	return
}

func (this *DeviceGroups) setGroupMetadata(taskId string, metadata messages.GroupTaskMetadata) (err error) {
	err = this.dbSet(METADATA_KEY_PREFIX+taskId, metadata)
	return
}

func (this *DeviceGroups) getSubResult(subTaskId string) (result interface{}, err error) {
	err = this.dbGet(RESULT_KEY_PREFIX+subTaskId, &result)
	return
}

func (this *DeviceGroups) setSubResult(subTaskId string, subResult interface{}) (err error) {
	err = this.dbSet(RESULT_KEY_PREFIX+subTaskId, subResult)
	return
}

var ErrNotFount = memcache.ErrCacheMiss

func (this *DeviceGroups) getTaskResults(taskId string) (metadata messages.GroupTaskMetadata, results []interface{}, missingSubTasks []messages.GroupTaskMetadataElement, err error) {
	metadata, err = this.getGroupMetadata(taskId)
	if err != nil {
		return metadata, nil, nil, err
	}
	for _, sub := range metadata.Children {
		subResult, err := this.getSubResult(sub.Task.Id)
		if err == ErrNotFount {
			err = nil
			missingSubTasks = append(missingSubTasks, sub)
		} else {
			results = append(results, subResult)
		}
	}
	return
}

func (this *DeviceGroups) GetSubTasks(request messages.Command, task messages.CamundaExternalTask) (result []messages.GroupTaskMetadataElement, err error) {
	if request.GroupId == "" {
		return []messages.GroupTaskMetadataElement{{
			Command: request,
			Task:    task,
		}}, nil
	}
	token, err := this.repo.GetToken(task.TenantId)
	if err != nil {
		return nil, err
	}
	group, err := this.repo.GetDeviceGroup(token, request.GroupId)
	if err != nil {
		return nil, err
	}
	for i, deviceId := range group.DeviceIds {
		device, err := this.repo.GetDevice(token, deviceId)
		if err != nil {
			return nil, err
		}

		deviceType, err := this.repo.GetDeviceType(token, device.DeviceTypeId)
		if err != nil {
			return nil, err
		}

		services := this.getFilteredServices(request, deviceType.Services)

		for j, service := range services {
			newTask := task
			newCommand := request

			newTask.Id = task.Id + "_" + strconv.Itoa(i) + "_" + strconv.Itoa(j)

			newCommand.DeviceId = deviceId
			newCommand.Device = &device

			newCommand.ServiceId = service.Id
			newCommand.Service = &service

			result = append(result, messages.GroupTaskMetadataElement{
				Command: newCommand,
				Task:    newTask,
			})
		}
	}
	return result, nil
}

func (this *DeviceGroups) dbGet(key string, value interface{}) error {
	item, err := this.db.Get(key)
	if err != nil {
		return err
	}
	err = json.Unmarshal(item.Value, value)
	return err
}

func (this *DeviceGroups) dbSet(key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return this.db.Set(&memcache.Item{Value: jsonValue, Expiration: this.expirationInSeconds, Key: key})
}

func (this *DeviceGroups) getFilteredServices(command messages.Command, services []model.Service) (result []model.Service) {
	serviceIndex := map[string]model.Service{}
	for _, service := range services {
		for _, functionId := range service.FunctionIds {
			if !(isMeasuringFunctionId(functionId) && service.Interaction == model.EVENT) { //mqtt cannot be measured in a task
				if functionId == command.Function.Id {
					if command.Aspect != nil {
						for _, aspect := range service.AspectIds {
							if aspect == command.Aspect.Id {
								serviceIndex[service.Id] = service
							}
						}
					}
					if command.DeviceClass != nil {
						serviceIndex[service.Id] = service
					}
				}
			}
		}
	}
	for _, service := range serviceIndex {
		result = append(result, service)
	}
	return result
}

func isMeasuringFunctionId(id string) bool {
	if strings.HasPrefix(id, model.MEASURING_FUNCTION_PREFIX) {
		return true
	}
	return false
}
