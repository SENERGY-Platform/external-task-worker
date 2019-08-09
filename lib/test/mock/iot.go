package mock

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/repo"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
)

type RepoMockFactory struct {
	devices map[string]model.DeviceInstance
	services map[string]model.Service
}

func NewRepoMock()  repo.FactoryInterface {
	return &RepoMockFactory{devices: map[string]model.DeviceInstance{}, services: map[string]model.Service{}}
}

func (this *RepoMockFactory) Get(configType util.ConfigType) repo.RepoInterface {
	return this
}

func (this *RepoMockFactory) GetDeviceInfo(instanceId string, serviceId string, user string) (instance model.DeviceInstance, service model.Service, err error) {
	instance, ok := this.devices[instanceId]
	if !ok {
		return instance, service, errors.New("device not found")
	}
	service, ok = this.services[instanceId]
	if !ok {
		return instance, service, errors.New("service not found")
	}
	return instance, service, nil
}

func (this *RepoMockFactory) RegisterDevice(device model.DeviceInstance){
	this.devices[device.Id] = device
}

func (this *RepoMockFactory) RegisterService(service model.Service){
	this.services[service.Id] = service
}