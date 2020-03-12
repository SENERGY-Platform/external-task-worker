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
	"encoding/json"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
	marshaller_service_configurables "github.com/SENERGY-Platform/marshaller/lib/configurables"
	marshaller_service "github.com/SENERGY-Platform/marshaller/lib/marshaller"
	marshaller_service_model "github.com/SENERGY-Platform/marshaller/lib/marshaller/model"
	"github.com/SENERGY-Platform/marshaller/lib/tests/mocks"
)

var Marshaller = &MarshallerMock{}

type MarshallerMock struct {
	marshaller *marshaller_service.Marshaller
}

func (this *MarshallerMock) New(url string) marshaller.Interface {
	this.marshaller = marshaller_service.New(mocks.Converter{}, mocks.ConceptRepo{})
	return this
}

func jsonCast(in interface{}, out interface{}) (err error) {
	temp, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(temp, out)
}

func (this *MarshallerMock) MarshalFromServiceAndProtocol(characteristicId string, service model.Service, protocol model.Protocol, characteristicData interface{}, configurables []marshaller.Configurable) (result map[string]string, err error) {
	mockService := marshaller_service_model.Service{}
	mockProtocol := marshaller_service_model.Protocol{}
	mockConfigurables := []marshaller_service_configurables.Configurable{}
	err = jsonCast(service, &mockService)
	if err != nil {
		return result, err
	}
	err = jsonCast(protocol, &mockProtocol)
	if err != nil {
		return result, err
	}
	err = jsonCast(configurables, &mockConfigurables)
	if err != nil {
		return result, err
	}
	return this.marshaller.MarshalInputs(mockProtocol, mockService, characteristicData, characteristicId, mockConfigurables...)
}

func (this *MarshallerMock) UnmarshalFromServiceAndProtocol(characteristicId string, service model.Service, protocol model.Protocol, message map[string]string, hints []string) (characteristicData interface{}, err error) {
	mockService := marshaller_service_model.Service{}
	mockProtocol := marshaller_service_model.Protocol{}
	err = jsonCast(service, &mockService)
	if err != nil {
		return characteristicData, err
	}
	err = jsonCast(protocol, &mockProtocol)
	if err != nil {
		return characteristicData, err
	}
	return this.marshaller.UnmarshalOutputs(mockProtocol, mockService, message, characteristicId, hints...)
}
