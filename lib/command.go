package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"runtime/debug"
	"strings"
)

func CreateCommandRequest(task messages.CamundaExternalTask) (command messages.Command, err error) {
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

func CreateCommandResult(msg messages.ProtocolMsg) (result messages.Command, err error) {
	result = messages.Command{
		DeviceId:         msg.Metadata.Device.Id,
		Device:           &msg.Metadata.Device,
		ServiceId:        msg.Metadata.Service.Id,
		Service:          &msg.Metadata.Service,
		Protocol:         &msg.Metadata.Protocol,
		ProtocolId:       msg.Metadata.Protocol.Id,
		CharacteristicId: msg.Metadata.OutputCharacteristic,
	}
	result.Output, err = marshaller.UnmarshalOutputs(*result.Protocol, *result.Service, msg.Response.Output, result.CharacteristicId)
	return
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
