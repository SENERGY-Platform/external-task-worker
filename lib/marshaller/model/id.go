package model

import "github.com/google/uuid"

func (variable *Characteristic) GenerateId() {
	variable.Id = URN_PREFIX + "characteristic:" + uuid.New().String()
	for i, v := range variable.SubCharacteristics {
		v.GenerateId()
		variable.SubCharacteristics[i] = v
	}
}

func (class *DeviceClass) GenerateId() {
	class.Id = URN_PREFIX + "device-class:" + uuid.New().String()
}

func (function *Function) GenerateId() {
	switch function.RdfType {
	case SES_ONTOLOGY_CONTROLLING_FUNCTION:
		function.Id = URN_PREFIX + "controlling-function:" + uuid.New().String()
	case SES_ONTOLOGY_MEASURING_FUNCTION:
		function.Id = URN_PREFIX + "measuring-function:" + uuid.New().String()
	default:
		function.Id = ""
	}
}

func (aspect *Aspect) GenerateId() {
	aspect.Id = URN_PREFIX + "aspect:" + uuid.New().String()
}

func (concept *Concept) GenerateId() {
	if concept.Id == "" {
		concept.Id = URN_PREFIX + "concept:" + uuid.New().String()
	}
	for i, characteristic := range concept.Characteristics {
		if characteristic.Id == "" {
			characteristic.GenerateId()
			concept.Characteristics[i] = characteristic
		}
	}
}

func (device *Device) GenerateId() {
	device.Id = URN_PREFIX + "device:" + uuid.New().String()
}

func (deviceType *DeviceType) GenerateId() {
	deviceType.Id = URN_PREFIX + "device-type:" + uuid.New().String()
	for i, service := range deviceType.Services {
		if service.Id == "" {
			service.GenerateId()
			deviceType.Services[i] = service
		}
	}
	if deviceType.DeviceClass.Id == "" {
		deviceType.DeviceClass.GenerateId()
	}
}

func (service *Service) GenerateId() {
	service.Id = URN_PREFIX + "service:" + uuid.New().String()
	for i, function := range service.Functions {
		if function.Id == "" {
			function.GenerateId()
			service.Functions[i] = function
		}
	}
	for i, aspect := range service.Aspects {
		if aspect.Id == "" {
			aspect.GenerateId()
			service.Aspects[i] = aspect
		}
	}
	for i, content := range service.Inputs {
		content.GenerateId()
		service.Inputs[i] = content
	}
	for i, content := range service.Outputs {
		content.GenerateId()
		service.Outputs[i] = content
	}
}

func (hub *Hub) GenerateId() {
	hub.Id = URN_PREFIX + "hub:" + uuid.New().String()
}

func (protocol *Protocol) GenerateId() {
	protocol.Id = URN_PREFIX + "protocol:" + uuid.New().String()
	for i, segment := range protocol.ProtocolSegments {
		segment.GenerateId()
		protocol.ProtocolSegments[i] = segment
	}
}

func (segment *ProtocolSegment) GenerateId() {
	segment.Id = URN_PREFIX + "protocol-segment:" + uuid.New().String()
}

func (content *Content) GenerateId() {
	content.Id = URN_PREFIX + "content:" + uuid.New().String()
	content.ContentVariable.GenerateId()
}

func (variable *ContentVariable) GenerateId() {
	variable.Id = URN_PREFIX + "content-variable:" + uuid.New().String()
	for i, v := range variable.SubContentVariables {
		v.GenerateId()
		variable.SubContentVariables[i] = v
	}
}
