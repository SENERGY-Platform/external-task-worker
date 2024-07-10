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

package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"runtime/debug"
	"strings"
)

func GetIncident(task messages.CamundaExternalTask) (incident *string) {
	payload, ok := task.Variables[util.CAMUNDA_VARIABLES_INCIDENT].Value.(string)
	if ok {
		return &payload
	}
	return nil
}

func GetCommandRequest(task messages.CamundaExternalTask) (command messages.Command, err error) {
	payload, ok := task.Variables[util.CAMUNDA_VARIABLES_PAYLOAD].Value.(string)
	if !ok {
		return command, errors.New(fmt.Sprint("ERROR: payload is not a string, ", task.Variables))
	}
	err = json.Unmarshal([]byte(payload), &command)
	if err != nil {
		return command, err
	}
	err = setDeviceOverwrite(&command, task)
	if err != nil {
		return command, err
	}
	parameter := getPayloadParameter(task)
	err = setPayloadParameter(&command, parameter)
	return
}

func setDeviceOverwrite(command *messages.Command, task messages.CamundaExternalTask) (err error) {
	comundaOverwriteVariableString, ok := task.Variables[util.CAMUNDA_VARIABLES_OVERWRITE]
	if ok {
		overwriteVariableString, ok := comundaOverwriteVariableString.Value.(string)
		if !ok {
			debug.PrintStack()
			return errors.New(util.CAMUNDA_VARIABLES_OVERWRITE + " variable is not a string")
		}
		overwriteVariable := messages.Overwrite{}
		err = json.Unmarshal([]byte(overwriteVariableString), &overwriteVariable)
		if err != nil {
			debug.PrintStack()
			return err
		}
		command.Service = overwriteVariable.Service
		command.Device = overwriteVariable.Device
		command.Protocol = overwriteVariable.Protocol
		command.ServiceId = overwriteVariable.ServiceId
		command.DeviceId = overwriteVariable.DeviceId
		command.ProtocolId = overwriteVariable.ProtocolId
	}
	return nil
}

func getPayloadParameter(task messages.CamundaExternalTask) (result map[string]interface{}) {
	result = map[string]interface{}{}
	for key, value := range task.Variables {
		path := strings.SplitN(key, ".", 2)
		if path[0] == "inputs" {
			var valueObj interface{}
			switch v := value.Value.(type) {
			case string:
				err := json.Unmarshal([]byte(v), &valueObj)
				if err != nil {
					err = nil
					valueObj = v
				}
			default:
				valueObj = v
			}

			if len(path) == 2 && path[1] != "" {
				result[path[1]] = valueObj
			} else {
				result[""] = valueObj
			}
		}
	}
	return
}

func setPayloadParameter(msg *messages.Command, parameter map[string]interface{}) (err error) {
	for paramName, value := range parameter {
		if paramName == "" {
			msg.Input, err = setVarOnPath(msg.Input, nil, value)
		} else {
			_, err = setVarOnPath(msg.Input, strings.Split(paramName, "."), value)
		}
		if err != nil {
			log.Println("ERROR: setPayloadParameter() -> ignore param", paramName, value, err)
			//return err
		}
	}
	return
}
