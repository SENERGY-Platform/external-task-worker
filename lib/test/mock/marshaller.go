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
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
	"github.com/SENERGY-Platform/marshaller/lib/config"
	marshaller_service_configurables "github.com/SENERGY-Platform/marshaller/lib/configurables"
	marshaller_service "github.com/SENERGY-Platform/marshaller/lib/marshaller"
	marshaller_service_model "github.com/SENERGY-Platform/marshaller/lib/marshaller/model"
	marshaller_service_v2 "github.com/SENERGY-Platform/marshaller/lib/marshaller/v2"
	"github.com/SENERGY-Platform/marshaller/lib/tests/mocks"
	"log"
)

var Marshaller = &MarshallerMock{}

type MarshallerMock struct {
	marshaller *marshaller_service.Marshaller
	v2         *marshaller_service_v2.Marshaller
}

func (this *MarshallerMock) New(ctx context.Context, url string) marshaller.Interface {
	conceptRepo, err := mocks.NewMockConceptRepo(ctx)
	if err != nil {
		panic(err)
	}
	this.marshaller = marshaller_service.New(mocks.Converter{}, conceptRepo, mocks.DeviceRepo)
	this.v2 = marshaller_service_v2.New(config.Config{}, mocks.Converter{}, conceptRepo)
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
	return this.marshaller.MarshalInputs(mockProtocol, mockService, characteristicData, characteristicId, nil, mockConfigurables...)
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
	return this.marshaller.UnmarshalOutputs(mockProtocol, mockService, message, characteristicId, nil, hints...)
}

func (this *MarshallerMock) MarshalV2(service model.Service, protocol model.Protocol, data []marshaller.MarshallingV2RequestData) (result map[string]string, err error) {
	mockService := marshaller_service_model.Service{}
	mockProtocol := marshaller_service_model.Protocol{}
	mockData := []marshaller_service_model.MarshallingV2RequestData{}
	err = jsonCast(service, &mockService)
	if err != nil {
		return result, err
	}
	err = jsonCast(protocol, &mockProtocol)
	if err != nil {
		return result, err
	}
	err = jsonCast(data, &mockData)
	if err != nil {
		return result, err
	}
	return this.v2.Marshal(mockProtocol, mockService, mockData)
}

func (this *MarshallerMock) UnmarshalV2(request marshaller.UnmarshallingV2Request) (result interface{}, err error) {
	mockProtocol := marshaller_service_model.Protocol{}
	err = jsonCast(request.Protocol, &mockProtocol)
	if err != nil {
		return result, err
	}
	mockService := marshaller_service_model.Service{}
	err = jsonCast(request.Service, &mockService)
	if err != nil {
		return result, err
	}
	var mockAspect *marshaller_service_model.AspectNode
	if request.AspectNode.Id != "" {
		err = jsonCast(request.AspectNode, mockAspect)
		if err != nil {
			return result, err
		}
	}
	if request.Path == "" {
		paths := this.v2.GetOutputPaths(mockService, request.FunctionId, mockAspect)
		if len(paths) > 0 {
			log.Println("WARNING: only first path found by FunctionId and AspectNode is used for Unmarshal")
		}
		if len(paths) == 0 {
			return result, errors.New("no output path found for criteria")
		}
		request.Path = paths[0]
	}
	return this.v2.Unmarshal(mockProtocol, mockService, request.CharacteristicId, request.Path, request.Message)
}
