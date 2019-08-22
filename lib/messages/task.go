package messages

import "github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"

type Command struct {
	//modeling time
	Function         model.Function `json:"function"`
	CharacteristicId string         `json:"characteristic_id"`

	//deployment time
	DeviceId   string         `json:"device_id,omitempty"`
	Device     model.Device   `json:"device,omitempty"`
	ServiceId  string         `json:"service_id,omitempty"`
	Service    model.Service  `json:"service,omitempty"`
	Protocol   model.Protocol `json:"protocol,omitempty"`
	ProtocolId string         `json:"protocol_id,omitempty"`

	//runtime
	Input  interface{} `json:"input,omitempty"`
	Output interface{} `json:"output,omitempty"`
}
