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
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"github.com/SENERGY-Platform/service-commons/pkg/cache"
	"github.com/SENERGY-Platform/service-commons/pkg/signal"
	"log"
	"net/url"
	"time"
)

type Iot struct {
	cache        *cache.Cache
	repoUrl      string
	keycloak     Keycloak
	cacheTimeout time.Duration
}

func NewIot(config util.Config) (*Iot, error) {
	c, err := cache.New(cache.Config{
		CacheInvalidationSignalHooks: map[cache.Signal]cache.ToKey{
			signal.Known.CacheInvalidationAll: nil,
			signal.Known.DeviceCacheInvalidation: func(signalValue string) (cacheKey string) {
				return "device." + signalValue
			},
			signal.Known.DeviceTypeCacheInvalidation: func(signalValue string) (cacheKey string) {
				return "device-type." + signalValue
			},
			signal.Known.DeviceGroupInvalidation: func(signalValue string) (cacheKey string) {
				return "device-group." + signalValue
			},
			signal.Known.ProtocolInvalidation: func(signalValue string) (cacheKey string) {
				return "protocol." + signalValue
			},
		},
	})
	if err != nil {
		return nil, err
	}
	cacheTimeout := time.Minute
	if config.CacheTimeout != "" && config.CacheTimeout != "-" {
		cacheTimeout, err = time.ParseDuration(config.CacheTimeout)
	}
	return &Iot{repoUrl: config.DeviceRepoUrl, cache: c, keycloak: Keycloak{config: config}, cacheTimeout: cacheTimeout}, err
}

func (this *Iot) GetToken(user string) (result Impersonate, err error) {
	return cache.UseWithExpInGet(this.cache, "user_token."+user, func() (Impersonate, time.Duration, error) {
		token, expirationInSec, err := this.keycloak.GetUserToken(user)
		return token, time.Duration(expirationInSec-10) * time.Second, err
	}, this.cacheTimeout)
}

func (this *Iot) GetDevice(token Impersonate, id string) (result model.Device, err error) {
	return cache.Use(this.cache, "device."+id, func() (model.Device, error) {
		return this.getDevice(token, id)
	}, this.cacheTimeout)
}

func (this *Iot) getDevice(token Impersonate, id string) (result model.Device, err error) {
	err = token.GetJSON(this.repoUrl+"/devices/"+url.QueryEscape(id), &result)
	return
}

func (this *Iot) GetProtocol(token Impersonate, id string) (result model.Protocol, err error) {
	return cache.Use(this.cache, "protocol."+id, func() (model.Protocol, error) {
		return this.getProtocol(token, id)
	}, this.cacheTimeout)
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
	var ok bool
	service, ok = item.(model.Service)
	if !ok {
		err = errors.New("unable to interpret cache value as model.Service")
	}
	return service, err
}

func (this *Iot) saveServiceToCache(service model.Service) {
	_ = this.cache.Set("service."+service.Id, service, this.cacheTimeout)
}

func (this *Iot) GetDeviceType(token Impersonate, id string) (result model.DeviceType, err error) {
	return cache.Use(this.cache, "device-type."+id, func() (model.DeviceType, error) {
		return this.getDeviceType(token, id)
	}, this.cacheTimeout)
}

func (this *Iot) getDeviceType(token Impersonate, id string) (result model.DeviceType, err error) {
	err = token.GetJSON(this.repoUrl+"/device-types/"+url.QueryEscape(id), &result)
	return
}

func (this *Iot) GetDeviceGroup(token Impersonate, id string) (result model.DeviceGroup, err error) {
	return cache.Use(this.cache, "device-group."+id, func() (model.DeviceGroup, error) {
		return this.getDeviceGroup(token, id)
	}, this.cacheTimeout)
}

func (this *Iot) getDeviceGroup(token Impersonate, id string) (result model.DeviceGroup, err error) {
	err = token.GetJSON(this.repoUrl+"/device-groups/"+url.QueryEscape(id), &result)
	return
}
