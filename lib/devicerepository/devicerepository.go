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
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
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

func (this *Iot) GetToken(user string) (result Impersonate, err error) {
	err = this.cache.UseWithExpiration("user_token."+user, func() (interface{}, int, error) {
		token, expirationInSec, err := this.keycloak.GetUserToken(user)
		return token, int(expirationInSec) - 10, err
	}, &result)
	return
}

func (this *Iot) GetDevice(token Impersonate, id string) (result model.Device, err error) {
	err = this.cache.Use("device."+id, func() (interface{}, error) {
		return this.getDevice(token, id)
	}, &result)
	return
}

func (this *Iot) getDevice(token Impersonate, id string) (result model.Device, err error) {
	err = token.GetJSON(this.repoUrl+"/devices/"+url.QueryEscape(id), &result)
	return
}

func (this *Iot) GetProtocol(token Impersonate, id string) (result model.Protocol, err error) {
	err = this.cache.Use("protocol."+id, func() (interface{}, error) {
		return this.getProtocol(token, id)
	}, &result)
	return
}

func (this *Iot) getProtocol(token Impersonate, id string) (result model.Protocol, err error) {
	err = token.GetJSON(this.repoUrl+"/protocols/"+url.QueryEscape(id), &result)
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

func (this *Iot) GetDeviceType(token Impersonate, id string) (result model.DeviceType, err error) {
	err = this.cache.Use("deviceType."+id, func() (interface{}, error) {
		return this.getDeviceType(token, id)
	}, &result)
	return
}

func (this *Iot) getDeviceType(token Impersonate, id string) (result model.DeviceType, err error) {
	err = token.GetJSON(this.repoUrl+"/device-types/"+url.QueryEscape(id), &result)
	return
}

func (this *Iot) GetDeviceGroup(token Impersonate, id string) (result model.DeviceGroup, err error) {
	err = this.cache.Use("deviceGroup."+id, func() (interface{}, error) {
		return this.getDeviceGroup(token, id)
	}, &result)
	return
}

func (this *Iot) getDeviceGroup(token Impersonate, id string) (result model.DeviceGroup, err error) {
	err = token.GetJSON(this.repoUrl+"/device-groups/"+url.QueryEscape(id), &result)
	return
}
