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

func (this *Marshaller) Unmarshal(characteristicId string, serviceId string, message map[string]string) (characteristicData interface{}, err error) {
	return SendUnmarshalRequest(this.url+"/unmarshal/"+url.PathEscape(serviceId)+"/"+url.PathEscape(characteristicId), UnmarshallingRequest{
		Message: message,
	})
}

func (this *Marshaller) UnmarshalFromService(characteristicId string, service model.Service, message map[string]string) (characteristicData interface{}, err error) {
	return SendUnmarshalRequest(this.url+"/unmarshal", UnmarshallingRequest{
		CharacteristicId: characteristicId,
		Service:          service,
		Message:          message,
	})
}

func (this *Marshaller) UnmarshalFromServiceAndProtocol(characteristicId string, service model.Service, protocol model.Protocol, message map[string]string) (characteristicData interface{}, err error) {
	return SendUnmarshalRequest(this.url, UnmarshallingRequest{
		CharacteristicId: characteristicId,
		Service:          service,
		Protocol:         &protocol,
		Message:          message,
	})
}

func SendUnmarshalRequest(url string, request UnmarshallingRequest) (characteristicData interface{}, err error) {
	body, err := json.Marshal(request)
	if err != nil {
		debug.PrintStack()
		return characteristicData, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		debug.PrintStack()
		return characteristicData, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		temp, _ := ioutil.ReadAll(resp.Body)
		debug.PrintStack()
		return characteristicData, errors.New(string(temp))
	}
	err = json.NewDecoder(resp.Body).Decode(&characteristicData)
	if err != nil {
		debug.PrintStack()
		return characteristicData, err
	}
	return characteristicData, err
}
