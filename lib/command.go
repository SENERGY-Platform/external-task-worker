package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"strings"
)

func CreateCommandRequest(task messages.CamundaTask) (request messages.Command, err error) {
	payload, ok := task.Variables[util.CAMUNDA_VARIABLES_PAYLOAD].Value.(string)
	if !ok {
		return request, errors.New(fmt.Sprint("ERROR: payload is not a string, ", task.Variables))
	}
	err = json.Unmarshal([]byte(payload), &request)
	if err != nil {
		return request, err
	}
	parameter := getPayloadParameter(task)
	err = setPayloadParameter(&request, parameter)
	return
}

func CreateCommandResult(msg messages.ProtocolMsg) (result messages.Command, err error) {
	result = messages.Command{
		DeviceId:         msg.Metadata.Device.Id,
		Device:           msg.Metadata.Device,
		ServiceId:        msg.Metadata.Service.Id,
		Service:          msg.Metadata.Service,
		Protocol:         msg.Metadata.Protocol,
		ProtocolId:       msg.Metadata.Protocol.Id,
		CharacteristicId: msg.Metadata.OutputCharacteristic,
	}
	result.Output, err = marshaller.UnmarshalOutputs(result.Protocol, result.Service, msg.Response.Output, result.CharacteristicId)
	return
}

func getPayloadParameter(task messages.CamundaTask) (result map[string]interface{}) {
	result = map[string]interface{}{}
	for key, value := range task.Variables {
		path := strings.SplitN(key, ".", 2)
		if len(path) == 2 && path[0] == "inputs" && path[1] != "" {
			result[path[1]] = value.Value
		}
	}
	return
}

func setPayloadParameter(msg *messages.Command, parameter map[string]interface{}) (err error) {
	for paramName, value := range parameter {
		_, err := setVarOnPath(msg.Input, strings.Split(paramName, "."), value)
		if err != nil {
			log.Println("ERROR: setPayloadParameter() -> ignore param", paramName, value, err)
			//return err
		}
	}
	return
}