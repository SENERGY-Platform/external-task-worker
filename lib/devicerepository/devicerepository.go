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

package devicerepository

import (
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"net/url"
)

type Iot struct {
	cache         *Cache
	repoUrl       string
	permsearchUrl string
	keycloak      Keycloak
}

func NewIot(config util.Config) *Iot {
	return &Iot{repoUrl: config.DeviceRepoUrl, cache: NewCache(), permsearchUrl: config.PermissionsUrl, keycloak: Keycloak{config: config}}
}

func (this *Iot) GetToken(user string) (Impersonate, error) {
	return this.keycloak.GetUserToken(user)
}

func (this *Iot) GetDevice(token Impersonate, id string) (result model.Device, err error) {
	result, err = this.getDeviceFromCache(id)
	if err != nil {
		err = token.GetJSON(this.repoUrl+"/devices/"+url.QueryEscape(id), &result)
		if err == nil {
			this.saveDeviceToCache(result)
		}
	}
	return
}

func (this *Iot) GetProtocol(token Impersonate, id string) (result model.Protocol, err error) {
	result, err = this.getProtocolFromCache(id)
	if err != nil {
		err = token.GetJSON(this.repoUrl+"/protocols/"+url.QueryEscape(id), &result)
		if err == nil {
			this.saveProtocolToCache(result)
		}
	}
	return
}

func (this *Iot) GetService(token Impersonate, device model.Device, id string) (result model.Service, err error) {
	result, err = this.getServiceFromCache(id)
	if err != nil {
		dt, err := this.GetDeviceType(token, device.DeviceTypeId)
		if err != nil {
			log.Println("ERROR: unable to load device-type", device.DeviceTypeId, token)
			return result, err
		}
		for _, service := range dt.Services {
			if service.Id == id {
				this.saveServiceToCache(service)
				return service, nil
			}
		}
		log.Println("ERROR: unable to find service in device-type", device.DeviceTypeId, id)
		return result, errors.New("service not found")
	}
	return
}

func (this *Iot) getDeviceFromCache(id string) (device model.Device, err error) {
	item, err := this.cache.Get("device." + id)
	if err != nil {
		return device, err
	}
	err = json.Unmarshal(item.Value, &device)
	return
}

func (this *Iot) saveDeviceToCache(device model.Device) {
	buffer, _ := json.Marshal(device)
	this.cache.Set("device."+device.Id, buffer)
}

func (this *Iot) getServiceFromCache(id string) (service model.Service, err error) {
	item, err := this.cache.Get("service." + id)
	if err != nil {
		return service, err
	}
	err = json.Unmarshal(item.Value, &service)
	return
}

func (this *Iot) saveServiceToCache(service model.Service) {
	buffer, _ := json.Marshal(service)
	this.cache.Set("service."+service.Id, buffer)
}

func (this *Iot) saveProtocolToCache(protocol model.Protocol) {
	buffer, _ := json.Marshal(protocol)
	this.cache.Set("protocol."+protocol.Id, buffer)
}

func (this *Iot) getProtocolFromCache(id string) (protocol model.Protocol, err error) {
	item, err := this.cache.Get("protocol." + id)
	if err != nil {
		return protocol, err
	}
	err = json.Unmarshal(item.Value, &protocol)
	return
}

func (this *Iot) GetDeviceType(token Impersonate, id string) (result model.DeviceType, err error) {
	result, err = this.getDeviceTypeFromCache(id)
	if err != nil {
		err = token.GetJSON(this.repoUrl+"/device-types/"+url.QueryEscape(id), &result)
	}
	if err == nil {
		this.saveDeviceTypeToCache(result)
	}
	return
}

func (this *Iot) saveDeviceTypeToCache(deviceType model.DeviceType) {
	buffer, _ := json.Marshal(deviceType)
	this.cache.Set("deviceType."+deviceType.Id, buffer)
}

func (this *Iot) getDeviceTypeFromCache(id string) (deviceType model.DeviceType, err error) {
	item, err := this.cache.Get("deviceType." + id)
	if err != nil {
		return deviceType, err
	}
	err = json.Unmarshal(item.Value, &deviceType)
	return
}
