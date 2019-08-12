package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	formatter_lib "github.com/SENERGY-Platform/formatter-lib"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
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

func CreateCommandResult(nrMsg messages.ProtocolMsg) (result messages.Command, err error) {
	result.Outputs = map[string]interface{}{}
	result.ServiceId = nrMsg.ServiceId
	service := nrMsg.Service
	for _, output := range nrMsg.ProtocolParts {
		for _, serviceOutput := range service.Output {
			if serviceOutput.MsgSegment.Name == output.Name {
				parsedOutput, err := formatter_lib.ParseFormat(serviceOutput.Type, serviceOutput.Format, output.Value, serviceOutput.AdditionalFormatinfo)
				if err != nil {
					log.Println("error on parsing")
					return result, err
				}
				outputInterface, err := formatter_lib.FormatToJsonStruct([]model.ConfigField{}, parsedOutput)
				if err != nil {
					return result, err
				}
				parsedOutput.Name = serviceOutput.Name
				result.Outputs[serviceOutput.Name] = outputInterface
			}
		}
	}
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
		_, err := setVarOnPath(msg.Inputs, strings.Split(paramName, "."), value)
		if err != nil {
			log.Println("ERROR: setPayloadParameter() -> ignore param", paramName, value, err)
			//return err
		}
	}
	return
}
