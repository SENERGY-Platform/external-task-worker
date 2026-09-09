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
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/marshaller/lib/api/messages"
)

func (this *Marshaller) MarshalV2(service model.Service, protocol model.Protocol, data []MarshallingV2RequestData) (result map[string]string, err error) {
	result, err, _ = this.client.MarshalV2(messages.MarshallingV2Request{
		Service:  service,
		Protocol: protocol,
		Data:     data,
	})
	return result, err
}

func (this *Marshaller) MarshalFromServiceAndProtocol(characteristicId string, service model.Service, protocol model.Protocol, characteristicData interface{}, configurables []Configurable) (result map[string]string, err error) {
	result, err, _ = this.client.Marshal(messages.MarshallingRequest{
		CharacteristicId: characteristicId,
		Service:          service,
		Protocol:         &protocol,
		Configurables:    configurables,
		Data:             characteristicData,
	})
	return result, err
}
