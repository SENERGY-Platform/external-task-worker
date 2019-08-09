package messages

import (
	"errors"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
)

type ProtocolPart struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ProtocolMsg struct {
	WorkerId           string         `json:"worker_id"`
	TaskId             string         `json:"task_id"`
	CompletionStrategy string         `json:"completion_strategy"`
	DeviceUrl          string         `json:"device_url"`
	ServiceUrl         string         `json:"service_url"`
	ProtocolParts      []ProtocolPart `json:"protocol_parts"`
	DeviceInstanceId   string         `json:"device_instance_id"`
	ServiceId          string         `json:"service_id"`
	OutputName         string         `json:"output_name"`
	Time               string         `json:"time"`
	Service            model.Service  `json:"service"`
}


type Envelope struct {
	DeviceId  string      `json:"device_id"`
	ServiceId string      `json:"service_id"`
	Value     interface{} `json:"value"`
}


func (envelope Envelope) Validate() error {
	if envelope.DeviceId == "" {
		return errors.New("missing device id")
	}
	if envelope.ServiceId == "" {
		return errors.New("missing service id")
	}
	return nil
}
