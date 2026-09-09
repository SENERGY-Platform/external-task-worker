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

func (this *Marshaller) UnmarshalV2(request UnmarshallingV2Request) (characteristicData interface{}, err error) {
	characteristicData, err, _ = this.client.UnmarshalV2(request)
	return characteristicData, err
}

func (this *Marshaller) UnmarshalFromServiceAndProtocol(characteristicId string, service model.Service, protocol model.Protocol, message map[string]string, hints []string) (characteristicData interface{}, err error) {
	characteristicData, err, _ = this.client.Unmarshal(messages.UnmarshallingRequest{
		CharacteristicId:     characteristicId,
		Service:              service,
		Protocol:             &protocol,
		Message:              message,
		ContentVariableHints: hints,
	})
	return characteristicData, err
}
