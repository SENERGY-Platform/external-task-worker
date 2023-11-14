/*
 * Copyright 2019 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mock

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/util"
)

var Repo = &RepoMock{ResetOnGetRepoInterface: true}

type RepoMock struct {
	ResetOnGetRepoInterface bool
	devices                 map[string]model.Device
	services                map[string]model.Service
	protocols               map[string]model.Protocol
	deviceTypes             map[string]model.DeviceType
	deviceGroups            map[string]model.DeviceGroup
}

func (this *RepoMock) GetDeviceType(token devicerepository.Impersonate, id string) (model.DeviceType, error) {
	dt, ok := this.deviceTypes[id]
	if !ok {
		return dt, errors.New("device-type not found")
	}
	return dt, nil
}

func (this *RepoMock) GetDeviceGroup(token devicerepository.Impersonate, id string) (model.DeviceGroup, error) {
	dg, ok := this.deviceGroups[id]
	if !ok {
		return dg, errors.New("device-group not found")
	}
	return dg, nil
}

func (this *RepoMock) Get(configType util.Config) (devicerepository.RepoInterface, error) {
	if this.ResetOnGetRepoInterface {
		this.devices = map[string]model.Device{}
		this.services = map[string]model.Service{}
		this.protocols = map[string]model.Protocol{}
		this.deviceTypes = map[string]model.DeviceType{}
		this.deviceGroups = map[string]model.DeviceGroup{}
	}
	return this, nil
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

func (this *RepoMock) RegisterDeviceType(deviceType model.DeviceType) {
	this.deviceTypes[deviceType.Id] = deviceType
	for _, s := range deviceType.Services {
		this.services[s.Id] = s
	}
}

func (this *RepoMock) RegisterDeviceGroup(deviceGroup model.DeviceGroup) {
	this.deviceGroups[deviceGroup.Id] = deviceGroup
}

func (this *RepoMock) New() *RepoMock {
	return &RepoMock{
		ResetOnGetRepoInterface: false,
		devices:                 map[string]model.Device{},
		services:                map[string]model.Service{},
		protocols:               map[string]model.Protocol{},
		deviceTypes:             map[string]model.DeviceType{},
		deviceGroups:            map[string]model.DeviceGroup{},
	}
}
