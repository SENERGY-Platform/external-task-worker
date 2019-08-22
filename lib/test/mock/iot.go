package mock

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"github.com/SENERGY-Platform/external-task-worker/util"
)

var Repo = &RepoMock{}

type RepoMock struct {
	devices  map[string]model.Device
	services map[string]model.Service
}

func (this *RepoMock) Get(configType util.Config) devicerepository.RepoInterface {
	this.devices = map[string]model.Device{}
	this.services = map[string]model.Service{}
	return this
}

func (this *RepoMock) GetDeviceInfo(instanceId string, serviceId string, user string) (instance model.DeviceInstance, service model.Service, err error) {
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

func (this *RepoMock) RegisterDevice(device model.DeviceInstance) {
	this.devices[device.Id] = device
}

func (this *RepoMock) RegisterService(service model.Service) {
	this.services[service.Id] = service
}