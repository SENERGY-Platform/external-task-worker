package messages

import "github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"

type TaskInfo struct {
	WorkerId            string `json:"worker_id"`
	TaskId              string `json:"task_id"`
	ProcessInstanceId   string `json:"process_instance_id"`
	ProcessDefinitionId string `json:"process_definition_id"`
	CompletionStrategy  string `json:"completion_strategy"`
	Time                string `json:"time"`
}

type ProtocolRequest struct {
	Input map[string]string `json:"input"`
}

type ProtocolResponse struct {
	Output map[string]string `json:"output"`
}

type Metadata struct {
	Device               model.Device   `json:"device"`
	Service              model.Service  `json:"service"`
	Protocol             model.Protocol `json:"protocol"`
	InputCharacteristic  string         `json:"input_characteristic,omitempty"`
	OutputCharacteristic string         `json:"output_characteristic,omitempty"`
}

type ProtocolMsg struct {
	Request  ProtocolRequest  `json:"request"`
	Response ProtocolResponse `json:"response"`
	TaskInfo TaskInfo         `json:"task_info"`
	Metadata Metadata         `json:"metadata"`
}
