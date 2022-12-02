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
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda/interfaces"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"github.com/bradfitz/gomemcache/memcache"
	"log"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Callback is implemented by: lib.CmdWorker::CreateProtocolMessage()
type Callback = func(command messages.Command, task messages.CamundaExternalTask) (request *messages.KafkaMessage, event *messages.EventRequest, err error)

func New(scheduler string, camunda interfaces.CamundaInterface, devicerepo devicerepository.RepoInterface, protocolMessageCallback Callback, currentlyRunningTimeoutInMs int64, expirationInSeconds int32, memcachedUrls []string, memcachedTimeout string, memcachedMaxIdleConns int64, filterEvents bool) *DeviceGroups {
	if len(memcachedUrls) == 0 {
		log.Println("start with local sub result storage")
		return NewWithKeyValueStore(scheduler, camunda, devicerepo, protocolMessageCallback, currentlyRunningTimeoutInMs, expirationInSeconds, NewLocalDb(), filterEvents)
	} else {
		client := memcache.New(memcachedUrls...)
		client.MaxIdleConns = int(memcachedMaxIdleConns)
		timeout, err := time.ParseDuration(memcachedTimeout)
		if err != nil {
			log.Println("WARNING: invalid memcached timeout; use default")
		} else {
			client.Timeout = timeout
		}
		return NewWithKeyValueStore(scheduler, camunda, devicerepo, protocolMessageCallback, currentlyRunningTimeoutInMs, expirationInSeconds, client, filterEvents)
	}
}

func NewWithKeyValueStore(scheduler string, camunda interfaces.CamundaInterface, devicerepo devicerepository.RepoInterface, protocolMessageCallback Callback, currentlyRunningTimeoutInMs int64, expirationInSeconds int32, db DbInterface, filterEvents bool) *DeviceGroups {
	return &DeviceGroups{
		protocolMessageCallback: protocolMessageCallback,
		repo:                    devicerepo,
		db:                      db,
		expirationInSeconds:     expirationInSeconds,
		camunda:                 camunda,
		scheduler:               scheduler,
		currentlyRunningTimeout: time.Duration(currentlyRunningTimeoutInMs) * time.Millisecond,
		filterEvents:            filterEvents,
	}
}

type DeviceGroups struct {
	protocolMessageCallback Callback
	repo                    devicerepository.RepoInterface
	db                      DbInterface
	expirationInSeconds     int32
	camunda                 interfaces.CamundaInterface
	scheduler               string
	currentlyRunningTimeout time.Duration
	filterEvents            bool
}

type RequestInfo = messages.RequestInfo

type RequestInfoList = messages.RequestInfoList

func (this *DeviceGroups) ProcessResponse(subTaskId string, subResult interface{}) (parent messages.GroupTaskMetadataElement, results []interface{}, finished bool, err error) {
	err = this.setSubResult(subTaskId, SubResultWrapper{Value: subResult})
	if err != nil {
		return parent, nil, false, err
	}
	parent, results, finished, err = this.isFinished(subTaskId)
	if err == nil && finished {
		this.clearTaskData(parent.Task.Id)
	}
	return
}

func (this *DeviceGroups) ProcessCommand(command messages.Command, task messages.CamundaExternalTask, caller string) (completed bool, nextMessages messages.RequestInfoList, finishedResults []interface{}, err error) {
	if this.shouldIgnoreTask(caller, task) {
		log.Println("DEBUG: ignore task execute call", caller, task)
		return false, nil, nil, nil
	}
	err = this.saveCallInfo(task, caller)
	if err != nil {
		log.Println("WARNING: unable to save last caller", err)
	}
	nextRequests, finishedResults, err := this.getNextRequests(command, task)
	if err != nil {
		return completed, nextRequests, finishedResults, err
	}
	switch this.scheduler {
	case util.SEQUENTIAL:
		nextRequests, err = this.annotateSubTaskStates(nextRequests)
		if err != nil {
			return completed, nextRequests, finishedResults, err
		}
		nextRequests = this.filterRetries(nextRequests, command.Retries)
		completed = len(nextRequests) == 0
		if completed {
			this.clearTaskData(task.Id)
			if len(finishedResults) == 0 {
				err = errors.New("unable to get any results for device-group")
			}
			return completed, nextRequests, finishedResults, err
		}
		nextRequests = this.filterCurrentlyRunning(nextRequests)
		if len(nextRequests) > 0 {
			nextRequests = nextRequests[:1] // possible place to implement batches in sequence
		}
		err = this.updateSubTaskState(nextRequests)
		return completed, nextRequests, finishedResults, err
	case util.ROUND_ROBIN:
		nextRequests, err = this.annotateSubTaskStates(nextRequests)
		if err != nil {
			return completed, nextRequests, finishedResults, err
		}
		nextRequests = this.filterRetries(nextRequests, command.Retries)
		completed = len(nextRequests) == 0
		if completed {
			this.clearTaskData(task.Id)
			if len(finishedResults) == 0 {
				err = errors.New("unable to get any results for device-group")
			}
			return completed, nextRequests, finishedResults, err
		}
		nextRequests = this.filterCurrentlyRunning(nextRequests)
		sort.Slice(nextRequests, func(i, j int) bool {
			return nextRequests[i].SubTaskState.TryCount < nextRequests[j].SubTaskState.TryCount
		})
		if len(nextRequests) > 0 {
			nextRequests = nextRequests[:1] // possible place to implement batches in sequence
		}
		err = this.updateSubTaskState(nextRequests)
		return completed, nextRequests, finishedResults, err
	case util.PARALLEL:
		noMoreRetries := command.Retries != -1 && command.Retries < task.Retries
		completed = len(nextRequests) == 0 || noMoreRetries
		if completed {
			this.clearTaskData(task.Id)
		}
		if noMoreRetries {
			nextRequests = RequestInfoList{}
			if len(finishedResults) == 0 {
				err = errors.New("unable to get any results for device-group")
			}
		} else {
			this.camunda.SetRetry(task.Id, task.TenantId, task.Retries+1)
		}
		return completed, nextRequests, finishedResults, err
	default:
		return completed, nextRequests, finishedResults, errors.New("unknown scheduler " + this.scheduler)
	}
}

func (this *DeviceGroups) getNextRequests(command messages.Command, task messages.CamundaExternalTask) (missingRequests RequestInfoList, finishedResults []interface{}, err error) {
	var missingSubTasks []messages.GroupTaskMetadataElement
	_, finishedResults, missingSubTasks, err = this.getTaskResults(task.Id)
	if err == ErrNotFount {
		err = nil
		missingSubTasks, err = this.GetSubTasks(command, task)
		if err != nil {
			return nil, nil, err
		}
		err = this.setGroupMetadata(task.Id, messages.GroupTaskMetadata{
			Parent: messages.GroupTaskMetadataElement{
				Command: command,
				Task:    task,
			},
			Children: missingSubTasks,
		})
		if err != nil {
			return nil, nil, err
		}
		for _, subTask := range missingSubTasks {
			err = this.setGroupMetadata(subTask.Task.Id, messages.GroupTaskMetadata{
				Parent: messages.GroupTaskMetadataElement{
					Command: command,
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
		request, event, err := this.protocolMessageCallback(subTask.Command, subTask.Task)
		if err != nil {
			return nil, nil, err
		}
		missingRequests = append(missingRequests, RequestInfo{
			Request:  request,
			Event:    event,
			Metadata: subTask,
		})
	}
	return
}

func (this *DeviceGroups) isFinished(taskId string) (parent messages.GroupTaskMetadataElement, results []interface{}, finished bool, err error) {
	meta, results, missing, err := this.getTaskResults(taskId)
	if err != nil {
		return parent, nil, false, err
	}
	finished = len(missing) == 0
	return meta.Parent, results, finished, err
}

func (this *DeviceGroups) getTaskResults(taskId string) (metadata messages.GroupTaskMetadata, results []interface{}, missingSubTasks []messages.GroupTaskMetadataElement, err error) {
	metadata, err = this.getGroupMetadata(taskId)
	if err != nil {
		return metadata, nil, nil, err
	}
	for _, sub := range metadata.Children {
		var subResult SubResultWrapper
		subResult, err = this.getSubResult(sub.Task.Id)
		if err == nil && !subResult.Failed {
			results = append(results, subResult.Value)
		}
		if err == ErrNotFount {
			err = nil
			missingSubTasks = append(missingSubTasks, sub)
		}
		if err != nil {
			return
		}
	}
	return
}

func (this *DeviceGroups) GetSubTasks(request messages.Command, task messages.CamundaExternalTask) (result []messages.GroupTaskMetadataElement, err error) {
	result, err = this.getSubTasks(request, task)
	if err != nil {
		return result, err
	}
	//remove id modifier parameter
	for i, e := range result {
		if e.Command.Device != nil {
			e.Command.Device.Id, _ = SplitModifier(e.Command.Device.Id)
			e.Command.Device.DeviceTypeId, _ = SplitModifier(e.Command.Device.DeviceTypeId)
		}
		if e.Command.DeviceId != "" {
			e.Command.DeviceId, _ = SplitModifier(e.Command.DeviceId)
		}
		result[i] = e
	}
	return result, nil
}

func (this *DeviceGroups) getSubTasks(request messages.Command, task messages.CamundaExternalTask) (result []messages.GroupTaskMetadataElement, err error) {
	if request.DeviceGroupId != "" {
		return this.GetGroupSubTasks(request, task)
	}
	if request.DeviceId != "" && request.ServiceId == "" {
		return this.GetEmptyServiceSubTasks(request, task)
	}
	return []messages.GroupTaskMetadataElement{{
		Command: request,
		Task:    task,
	}}, nil
}

func (this *DeviceGroups) GetEmptyServiceSubTasks(request messages.Command, task messages.CamundaExternalTask) (result []messages.GroupTaskMetadataElement, err error) {
	token, err := this.repo.GetToken(task.TenantId)
	if err != nil {
		return nil, err
	}
	device, err := this.repo.GetDevice(token, request.DeviceId)
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

		newTask.Id = task.Id + "_0_" + strconv.Itoa(j)

		newCommand.DeviceId = request.DeviceId
		newCommand.Device = &device

		newCommand.ServiceId = service.Id
		temp := service
		newCommand.Service = &temp

		result = append(result, messages.GroupTaskMetadataElement{
			Command: newCommand,
			Task:    newTask,
		})
	}
	return result, nil
}

func (this *DeviceGroups) GetGroupSubTasks(request messages.Command, task messages.CamundaExternalTask) (result []messages.GroupTaskMetadataElement, err error) {
	if request.DeviceGroupId == "" {
		return []messages.GroupTaskMetadataElement{{
			Command: request,
			Task:    task,
		}}, nil
	}
	token, err := this.repo.GetToken(task.TenantId)
	if err != nil {
		return nil, err
	}
	group, err := this.repo.GetDeviceGroup(token, request.DeviceGroupId)
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
			temp := service
			newCommand.Service = &temp

			result = append(result, messages.GroupTaskMetadataElement{
				Command: newCommand,
				Task:    newTask,
			})
		}
	}
	return result, nil
}

func (this *DeviceGroups) getFilteredServices(command messages.Command, services []model.Service) (result []model.Service) {
	serviceIndex := map[string]model.Service{}
	for _, service := range services {
		contents := service.Inputs
		if isMeasuringFunctionId(command.Function.Id) {
			contents = service.Outputs
		}
		aspect := model.AspectNode{}
		if command.Aspect != nil {
			aspect = *command.Aspect
		}
		matchesCriteria := anyContentMatchesCriteria(contents, model.DeviceGroupFilterCriteria{FunctionId: command.Function.Id, AspectId: aspect.Id}, aspect)
		isEvent := isMeasuringFunctionId(command.Function.Id) && service.Interaction == model.EVENT
		if matchesCriteria && (!this.filterEvents || !isEvent) {
			serviceIndex[service.Id] = service
		}
	}
	for _, service := range serviceIndex {
		result = append(result, service)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Id < result[j].Id
	})
	return result
}

func anyContentMatchesCriteria(contents []model.Content, criteria model.DeviceGroupFilterCriteria, aspectNode model.AspectNode) bool {
	for _, content := range contents {
		if contentVariableContainsCriteria(content.ContentVariable, criteria, aspectNode) {
			return true
		}
	}
	return false
}

func contentVariableContainsCriteria(variable model.ContentVariable, criteria model.DeviceGroupFilterCriteria, aspectNode model.AspectNode) bool {
	if variable.FunctionId == criteria.FunctionId &&
		(criteria.AspectId == "" ||
			variable.AspectId == criteria.AspectId ||
			listContains(aspectNode.DescendentIds, variable.AspectId)) {
		return true
	}
	for _, sub := range variable.SubContentVariables {
		if contentVariableContainsCriteria(sub, criteria, aspectNode) {
			return true
		}
	}
	return false
}

func listContains(list []string, search string) bool {
	for _, element := range list {
		if element == search {
			return true
		}
	}
	return false
}

func (this *DeviceGroups) clearTaskData(parentTaskId string) {
	meta, err := this.getGroupMetadata(parentTaskId)
	if err == ErrNotFount {
		err = nil
		return
	}
	if err != nil {
		log.Println("WARNING: unable to delete data for", parentTaskId, err)
		return
	}
	elements := append(meta.Children, meta.Parent)
	for _, element := range elements {
		_ = this.db.Delete(METADATA_KEY_PREFIX + element.Task.Id)
		_ = this.db.Delete(RESULT_KEY_PREFIX + element.Task.Id)
		_ = this.db.Delete(LAST_CALL_INFO + element.Task.Id)
	}
}

func isMeasuringFunctionId(id string) bool {
	if strings.HasPrefix(id, model.MEASURING_FUNCTION_PREFIX) {
		return true
	}
	return false
}

func (this *DeviceGroups) shouldIgnoreTask(caller string, task messages.CamundaExternalTask) bool {
	if this.scheduler == util.PARALLEL {
		return caller == util.CALLER_RESPONSE
	}
	if caller == util.CALLER_RESPONSE {
		return false
	}
	lastCaller, lastCallDuration, err := this.getLastCallInfo(task)
	if err == ErrNotFount {
		//is first call
		return false
	}
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return true
	}
	if lastCaller == util.CALLER_CAMUNDA_LOOP {
		return false
	}
	return lastCallDuration < this.currentlyRunningTimeout
}

func (this *DeviceGroups) saveCallInfo(task messages.CamundaExternalTask, caller string) error {
	if this.scheduler == util.PARALLEL {
		return nil
	}
	return this.dbSet(LAST_CALL_INFO+task.Id, LastCallInfo{
		Caller: caller,
		Time:   time.Now(),
	})
}

func (this *DeviceGroups) getLastCallInfo(task messages.CamundaExternalTask) (lastCaller string, lastCallDuration time.Duration, err error) {
	lastCall := LastCallInfo{}
	err = this.dbGet(LAST_CALL_INFO+task.Id, &lastCall)
	lastCaller = lastCall.Caller
	lastCallDuration = time.Since(lastCall.Time)
	return
}

type LastCallInfo struct {
	Caller string    `json:"caller"`
	Time   time.Time `json:"time"`
}
