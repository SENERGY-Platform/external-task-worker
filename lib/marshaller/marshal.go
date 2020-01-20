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

package marshaller

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime/debug"
)

func (this *Marshaller) Marshal(characteristicId string, serviceId string, characteristicData interface{}, configurables []Configurable) (result map[string]string, err error) {
	return SendMarshalRequest(this.url+"/marshal/"+url.PathEscape(serviceId)+"/"+url.PathEscape(characteristicId), MarshallingRequest{
		Configurables: configurables,
		Data:          characteristicData,
	})
}

func (this *Marshaller) MarshalFromService(characteristicId string, service model.Service, characteristicData interface{}, configurables []Configurable) (result map[string]string, err error) {
	return SendMarshalRequest(this.url+"/marshal", MarshallingRequest{
		CharacteristicId: characteristicId,
		Service:          service,
		Configurables:    configurables,
		Data:             characteristicData,
	})
}

func (this *Marshaller) MarshalFromServiceAndProtocol(characteristicId string, service model.Service, protocol model.Protocol, characteristicData interface{}, configurables []Configurable) (result map[string]string, err error) {
	return SendMarshalRequest(this.url+"/marshal", MarshallingRequest{
		CharacteristicId: characteristicId,
		Service:          service,
		Protocol:         &protocol,
		Configurables:    configurables,
		Data:             characteristicData,
	})
}

func SendMarshalRequest(url string, request MarshallingRequest) (result map[string]string, err error) {
	body, err := json.Marshal(request)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		temp, _ := ioutil.ReadAll(resp.Body)
		debug.PrintStack()
		return result, errors.New(string(temp))
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	return result, err
}
