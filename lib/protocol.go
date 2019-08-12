package lib

import (
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	formatter_lib "github.com/SENERGY-Platform/formatter-lib"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"log"
	"strconv"
	"time"
)


func (this *worker) CreateProtocolMessage(request messages.SenergyTask, task messages.CamundaTask) (protocolTopic string, message string, err error) {
	instance, service, err := this.repository.GetDeviceInfo(request.InstanceId, request.ServiceId, task.TenantId)
	if err != nil {
		log.Println("error on CreateProtocolMessage getDeviceInfo: ", err)
		err = errors.New("unable to find device or service")
		return
	}
	value, err := this.createMessageForProtocolHandler(instance, service, request.Inputs, task)
	if err != nil {
		log.Println("ERROR: on CreateProtocolMessage createMessageForProtocolHandler(): ", err)
		err = errors.New("internal format error (inconsistent data?) (time: " + time.Now().String() + ")")
		return
	}
	protocolTopic = service.Protocol.ProtocolHandlerUrl
	if protocolTopic == "" {
		log.Println("ERROR: empty protocol topic")
		log.Println("DEBUG: ", instance, service)
		err = errors.New("empty protocol topic")
		return
	}
	envelope := messages.Envelope{ServiceId: service.Id, DeviceId: instance.Id, Value: value}
	if err := envelope.Validate(); err != nil {
		return protocolTopic, message, err
	}
	msg, err := json.Marshal(envelope)
	return protocolTopic, string(msg), err
}

func (this *worker) createMessageForProtocolHandler(instance model.DeviceInstance, service model.Service, inputs map[string]interface{}, task messages.CamundaTask) (result messages.ProtocolMsg, err error) {
	result = messages.ProtocolMsg{
		WorkerId:           this.camunda.GetWorkerId(),
		CompletionStrategy: util.Config.CompletionStrategy,
		DeviceUrl:          instance.Url,
		ServiceUrl:         service.Url,
		TaskId:             task.Id,
		DeviceInstanceId:   instance.Id,
		ServiceId:          service.Id,
		OutputName:         "result", //task.ActivityId,
		Time:               strconv.FormatInt(time.Now().Unix(), 10),
		Service:            service,
	}
	for _, serviceInput := range service.Input {
		for name, inputInterface := range inputs {
			if serviceInput.Name == name {
				input, err := formatter_lib.ParseFromJsonInterface(serviceInput.Type, inputInterface)
				if err != nil {
					return result, err
				}
				input.Name = name
				if err := formatter_lib.UseLiterals(&input, serviceInput.Type); err != nil {
					return result, err
				}
				formatedInput, err := formatter_lib.GetFormatedValue(instance.Config, serviceInput.Format, input, serviceInput.AdditionalFormatinfo)
				if err != nil {
					return result, err
				}
				result.ProtocolParts = append(result.ProtocolParts, messages.ProtocolPart{
					Name:  serviceInput.MsgSegment.Name,
					Value: formatedInput,
				})
			}
		}
	}
	return
}
