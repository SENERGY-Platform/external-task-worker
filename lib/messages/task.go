package messages


type Command struct {
	InstanceId string                 `json:"instance_id,omitempty"`
	ServiceId  string                 `json:"service_id,omitempty"`
	Inputs     map[string]interface{} `json:"inputs,omitempty"`
	Outputs    map[string]interface{} `json:"outputs,omitempty"`
	ErrorMsg   string                 `json:"error_msg,omitempty"`
}

