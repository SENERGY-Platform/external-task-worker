/*
 * Copyright 2022 InfAI (CC SES)
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

package lib

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/lib/timescale"
	marshallermodel "github.com/SENERGY-Platform/marshaller/lib/marshaller/model"
	"github.com/SENERGY-Platform/marshaller/lib/marshaller/serialization"
	"github.com/SENERGY-Platform/models/go/models"
)

func (this *CmdWorker) GetLastEventValue(token string, userId string, device model.Device, service model.Service, protocol model.Protocol, characteristicId string, functionId string, aspect model.AspectNode, timeout time.Duration) (code int, result interface{}) {
	output, err, code := this.getLastEventMessage(token, device, service, protocol, timeout)
	if err != nil {
		this.config.GetLogger().Error("unable to get event value:", "error", err)
		return 500, "unable to get event value: " + err.Error()
	}
	if code >= 300 {
		this.config.GetLogger().Error("unexpected getLastEventMessage() return code", "code", code)
		return code, "unexpected code " + strconv.Itoa(code)
	}
	marshalStartTime := time.Now()
	temp, err := this.marshaller.UnmarshalV2(marshaller.UnmarshallingV2Request{
		Service:          service,
		Protocol:         protocol,
		CharacteristicId: characteristicId,
		Message:          output,
		FunctionId:       functionId,
		AspectNode:       aspect,
		AspectNodeId:     aspect.Id,
	})
	if err != nil {
		//log marshal latency
		marshalDuration := time.Since(marshalStartTime)
		this.metrics.LogTaskMarshallingLatency("UnmarshalV2", userId, service.Id, functionId, marshalDuration)

		if this.config.Debug {
			marshalRequestStr, _ := json.Marshal(marshaller.UnmarshallingV2Request{
				Service:          service,
				Protocol:         protocol,
				CharacteristicId: characteristicId,
				Message:          output,
				FunctionId:       functionId,
				AspectNode:       aspect,
				AspectNodeId:     aspect.Id,
			})
			this.config.GetLogger().Debug("unable to unmarshal event value", "error", err, "request", string(marshalRequestStr))
		}
		this.config.GetLogger().Error("unable to unmarshal event value", "error", err)
		return http.StatusInternalServerError, "unable to unmarshal event value: " + err.Error()
	}
	return 200, temp
}

func (this *CmdWorker) getLastEventMessage(token string, device model.Device, service model.Service, protocol model.Protocol, timeout time.Duration) (result map[string]string, err error, code int) {
	request := createTimescaleRequest(device, service)
	response := []messages.TimescaleResponse{}
	response, err = this.timescale.Query(token, request, timeout)
	this.config.GetLogger().Debug("getLastEventMessage()", "request", timescale.CastRequest(request), "response", response, "error", err)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if len(request) != len(response) {
		return result, errors.New("timescale response has less elements then the request"), http.StatusInternalServerError
	}
	result, err, code = this.createEventValueFromTimescaleValues(service, protocol, request, response)
	return result, err, code
}

func (this *CmdWorker) createEventValueFromTimescaleValues(service model.Service, protocol model.Protocol, request []messages.TimescaleRequest, response []messages.TimescaleResponse) (result map[string]string, err error, code int) {
	result = map[string]string{}
	timescaleValue, err, code := this.getTimescaleValue(request, response)
	if err != nil {
		return result, err, code
	}
	for _, content := range service.Outputs {
		segmentValue := timescaleValue[content.ContentVariable.Name]
		segmentName := getSegmentName(protocol, content.ProtocolSegmentId)
		if segmentName != "" && segmentValue != nil {
			result[segmentName], err = marshalSegmentValue(content.Serialization, segmentValue, content.ContentVariable.Name)
			if err != nil {
				this.config.GetLogger().Error("unable to marshal segment value", "error", err)
				return result, err, http.StatusInternalServerError
			}
		}
	}
	return result, err, http.StatusOK
}

func marshalSegmentValue(serializationTo models.Serialization, value interface{}, rootName string) (string, error) {
	m, ok := serialization.Get(serializationTo)
	if !ok {
		return "", errors.New("unknown serialization")
	}
	return m.Marshal(value, marshallermodel.ContentVariable{Name: rootName})
}

var ErrMissingLastValue = errors.New("missing last value in mgw-last-value")
var ErrMissingLastValueCode = 513 //custom code to signify missing last-value in mgw-last-value

func (this *CmdWorker) getTimescaleValue(timescaleRequests []messages.TimescaleRequest, timescaleResponses []messages.TimescaleResponse) (result map[string]interface{}, err error, code int) {
	pathToValue := map[string]interface{}{}
	paths := []string{}
	for i, request := range timescaleRequests {
		if this.config.HandleMissingLastValueTimeAsError && timescaleResponses[i].Time == nil {
			return result, ErrMissingLastValue, ErrMissingLastValueCode
		}
		pathToValue[request.ColumnName] = timescaleResponses[i].Value
		paths = append(paths, request.ColumnName)
	}

	sort.Strings(paths)

	result = map[string]interface{}{}
	for _, path := range paths {
		result = setPath(result, strings.Split(path, "."), pathToValue[path])
	}

	return result, nil, http.StatusOK
}

func setPath(orig map[string]interface{}, path []string, value interface{}) map[string]interface{} {
	if len(path) == 0 {
		return orig
	}
	first := path[0]
	if len(path) == 1 {
		orig[first] = value
		return orig
	}
	rest := path[1:]
	sub, ok := orig[first]
	if !ok {
		orig[first] = setPath(map[string]interface{}{}, rest, value)
	} else {
		subMap, okCast := sub.(map[string]interface{})
		if !okCast {
			slog.Default().Error("setPath() expect map in path", "orig", orig, "sub", sub, "path", strings.Join(path, "."))
			return orig
		}
		orig[first] = setPath(subMap, rest, value)
	}
	return orig
}

func getSegmentName(protocol model.Protocol, id string) string {
	for _, segment := range protocol.ProtocolSegments {
		if segment.Id == id {
			return segment.Name
		}
	}
	return ""
}

func createTimescaleRequest(device model.Device, service model.Service) (result []messages.TimescaleRequest) {
	paths := []string{}
	for _, content := range service.Outputs {
		paths = append(paths, getContentPaths([]string{}, content.ContentVariable)...)
	}
	for _, path := range paths {
		result = append(result, messages.TimescaleRequest{
			Device:     device,
			Service:    service,
			ColumnName: path,
		})
	}
	return result
}

func getContentPaths(current []string, variable model.ContentVariable) (result []string) {
	//skip empty content
	if variable.Name == "" {
		return result
	}
	//skip list
	if variable.Type == model.List {
		return result
	}
	if len(variable.SubContentVariables) == 0 {
		return []string{strings.Join(append(current, variable.Name), ".")}
	}
	for _, sub := range variable.SubContentVariables {
		result = append(result, getContentPaths(append(current, variable.Name), sub)...)
	}
	return result
}
