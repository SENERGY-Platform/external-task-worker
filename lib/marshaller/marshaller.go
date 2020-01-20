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

import "github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"

type Marshaller struct {
	url string
}

func New(url string) *Marshaller {
	return &Marshaller{url: url}
}

type Interface interface {
	MarshalFromServiceAndProtocol(characteristicId string, service model.Service, protocol model.Protocol, characteristicData interface{}, configurables []Configurable) (result map[string]string, err error)
	UnmarshalFromServiceAndProtocol(characteristicId string, service model.Service, protocol model.Protocol, message map[string]string) (characteristicData interface{}, err error)
}

type FactoryInterface interface {
	New(url string) Interface
}

type MarshallerFactory struct{}

func (this MarshallerFactory) New(url string) Interface {
	return New(url)
}

var Factory = MarshallerFactory{}
