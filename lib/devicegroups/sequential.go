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
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"time"
)

type MessageWithState struct {
	Message messages.KafkaMessage
	State   SubTaskState
}

func (this *DeviceGroups) annotateSubTaskStates(requests RequestInfoList) (result RequestInfoList, err error) {
	for _, element := range requests {
		state, err := this.getSubTaskState(element.Metadata.Task.Id)
		if err != ErrNotFount && err != nil {
			return result, err
		}
		err = nil
		element.SubTaskState = state
		result = append(result, element)
	}
	return
}

func (this *DeviceGroups) filterRetries(nextRequests RequestInfoList, retries int64) (filteredRequests RequestInfoList) {
	for _, element := range nextRequests {
		if retries == -1 || element.SubTaskState.TryCount <= retries {
			filteredRequests = append(filteredRequests, element)
		}
	}
	return
}

func (this *DeviceGroups) updateSubTaskState(nextRequests RequestInfoList) (err error) {
	for _, element := range nextRequests {
		err = this.setSubTaskState(element.Metadata.Task.Id, SubTaskState{
			LastTry:  time.Now(),
			TryCount: element.SubTaskState.TryCount + 1,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *DeviceGroups) filterCurrentlyRunning(nextRequests RequestInfoList) (filteredRequests RequestInfoList) {
	for _, element := range nextRequests {
		if time.Since(element.SubTaskState.LastTry) > this.currentlyRunningTimeout {
			filteredRequests = append(filteredRequests, element)
		}
	}
	return
}
