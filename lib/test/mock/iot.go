package mock

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"github.com/SENERGY-Platform/external-task-worker/util"
)

var Repo = &RepoMock{}

type RepoMock struct {
	devices   map[string]model.Device
	services  map[string]model.Service
	protocols map[string]model.Protocol
}

func (this *RepoMock) Get(configType util.Config) devicerepository.RepoInterface {
	this.devices = map[string]model.Device{}
	this.services = map[string]model.Service{}
	this.protocols = map[string]model.Protocol{}
	return this
}

func (this *RepoMock) GetDevice(token devicerepository.Impersonate, id string) (model.Device, error) {
	device, ok := this.devices[id]
	if !ok {
		return device, errors.New("device not found")
	}
	return device, nil
}

func (this *RepoMock) GetService(token devicerepository.Impersonate, device model.Device, serviceId string) (model.Service, error) {
	service, ok := this.services[serviceId]
	if !ok {
		return service, errors.New("service not found")
	}
	return service, nil
}

func (this *RepoMock) GetProtocol(token devicerepository.Impersonate, id string) (model.Protocol, error) {
	protocol, ok := this.protocols[id]
	if !ok {
		return protocol, errors.New("protocol not found")
	}
	return protocol, nil
}

func (this *RepoMock) GetToken(user string) (devicerepository.Impersonate, error) {
	return "", nil
}

func (this *RepoMock) GetDeviceInfo(instanceId string, serviceId string, user string) (instance model.Device, service model.Service, err error) {
	instance, ok := this.devices[instanceId]
	if !ok {
		return instance, service, errors.New("device not found")
	}
	service, ok = this.services[serviceId]
	if !ok {
		return instance, service, errors.New("service not found")
	}
	return instance, service, nil
}

func (this *RepoMock) RegisterDevice(device model.Device) {
	this.devices[device.Id] = device
}

func (this *RepoMock) RegisterService(service model.Service) {
	this.services[service.Id] = service
}

func (this *RepoMock) RegisterProtocol(protocol model.Protocol) {
	this.protocols[protocol.Id] = protocol
}
