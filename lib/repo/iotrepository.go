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

package repo

import (
	"encoding/json"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"log"
	"net/url"

	"errors"
)

type Iot struct {
	cache         *Cache
	repoUrl       string
	permsearchUrl string
	keycloak Keycloak
}

func NewIot(config util.Config) (*Iot){
	return &Iot{repoUrl: config.DeviceRepoUrl, cache:NewCache(), permsearchUrl: config.PermissionsUrl, keycloak:Keycloak{config:config}}
}

func (this *Iot) GetDeviceInstance(token Impersonate, deviceInstanceId string) (result model.DeviceInstance, err error) {
	if err = this.CheckExecutionAccess(token, deviceInstanceId); err == nil {
		result, err = this.getDeviceFromCache(deviceInstanceId)
		if err != nil {
			err = token.GetJSON(this.repoUrl+"/devices/"+url.QueryEscape(deviceInstanceId), &result)
		}
	}
	return
}

func (this *Iot) getDeviceFromCache(id string) (device model.DeviceInstance, err error) {
	item, err := this.cache.Get("device."+id)
	if err != nil {
		return device, err
	}
	err = json.Unmarshal(item.Value, &device)
	return
}

func (this *Iot) getServiceFromCache(id string) (service model.Service, err error) {
	item, err := this.cache.Get("service."+id)
	if err != nil {
		return service, err
	}
	err = json.Unmarshal(item.Value, &service)
	return
}

func (this *Iot) GetDeviceService(token Impersonate, serviceId string) (result model.Service, err error) {
	result, err = this.getServiceFromCache(serviceId)
	if err != nil {
		err = token.GetJSON(this.repoUrl+"/services/"+url.QueryEscape(serviceId), &result)
	}
	return
}

func (this *Iot) CheckExecutionAccess(token Impersonate, deviceId string) (err error) {
	result, err := this.getAccessFromCache(token, deviceId)
	if err != nil {
		err = token.GetJSON(this.permsearchUrl + "/jwt/check/deviceinstance/" + url.QueryEscape(deviceId) + "/x/bool", &result)
	}
	if err != nil {
		return err
	}
	if result {
		return nil
	}else{
		return errors.New("user may not execute events for the resource")
	}
}

func (this *Iot) getAccessFromCache(token Impersonate, id string) (result bool, err error) {
	item, err := this.cache.Get("check.device."+string(token)+"."+id)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(item.Value, &result)
	return
}

func (this *Iot)  GetDeviceInfo(instanceId string, serviceId string, user string) (instance model.DeviceInstance, service model.Service, err error) {
	token, err := this.keycloak.GetUserToken(user)
	if err != nil {
		log.Println("error on user token generation: ", err)
		return instance, service, err
	}
	instance, err = this.GetDeviceInstance(token, instanceId)
	if err != nil {
		log.Println("error on getDeviceInfo GetDeviceInstance")
		return
	}
	service, err = this.GetDeviceService(token, serviceId)
	return
}
