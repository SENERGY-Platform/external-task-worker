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

package timescale

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"io"
	"net/http"
	"strings"
	"time"
)

type Interface interface {
	Query(token string, request []messages.TimescaleRequest, timeout time.Duration) (result []messages.TimescaleResponse, err error)
}

type Timescale struct {
	TimescaleWrapperUrl string
}

func Factory(ctx context.Context, config util.Config) (Interface, error) {
	return NewTimescale(config.TimescaleWrapperUrl), nil
}

type FactoryInterface func(ctx context.Context, config util.Config) (Interface, error)

func NewTimescale(timescaleWrapperUrl string) *Timescale {
	return &Timescale{TimescaleWrapperUrl: timescaleWrapperUrl}
}

func (this *Timescale) Query(token string, request []messages.TimescaleRequest, timeout time.Duration) (result []messages.TimescaleResponse, err error) {
	body := &bytes.Buffer{}
	query := CastRequest(request)
	err = json.NewEncoder(body).Encode(query)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", this.TimescaleWrapperUrl+"/queries", body)
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", token)
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		return result, errors.New(strings.TrimSpace(string(temp)))
	}
	queryResponse := TimescaleQueryResponse{}
	err = json.NewDecoder(resp.Body).Decode(&queryResponse)
	if err != nil {
		return result, err
	}
	return castQueryResponse(request, query, queryResponse)
}

func CastRequest(request []messages.TimescaleRequest) (result TimescaleQuery) {
	groupedRequests := map[string]map[string][]string{}
	for _, req := range request {
		if _, ok := groupedRequests[req.Device.Id]; !ok {
			groupedRequests[req.Device.Id] = map[string][]string{}
		}
		groupedRequests[req.Device.Id][req.Service.Id] = append(groupedRequests[req.Device.Id][req.Service.Id], req.ColumnName)
	}
	for deviceId, groupedByService := range groupedRequests {
		for serviceId, columns := range groupedByService {
			element := QueriesRequestElement{
				DeviceId:  deviceId,
				ServiceId: serviceId,
				Limit:     1,
			}
			for _, column := range columns {
				element.Columns = append(element.Columns, QueriesRequestElementColumn{
					Name: column,
				})
			}
			result = append(result, element)
		}
	}
	return
}

func castQueryResponse(request []messages.TimescaleRequest, query TimescaleQuery, response TimescaleQueryResponse) (result []messages.TimescaleResponse, err error) {
	deviceServiceToIndex := map[string]map[string]int{}
	deviceServiceColumnToIndex := map[string]map[string]map[string]int{}
	for i, q := range query {
		if _, ok := deviceServiceToIndex[q.DeviceId]; !ok {
			deviceServiceToIndex[q.DeviceId] = map[string]int{}
		}
		if _, ok := deviceServiceColumnToIndex[q.DeviceId]; !ok {
			deviceServiceColumnToIndex[q.DeviceId] = map[string]map[string]int{}
		}
		deviceServiceToIndex[q.DeviceId][q.ServiceId] = i
		if _, ok := deviceServiceColumnToIndex[q.DeviceId][q.ServiceId]; !ok {
			deviceServiceColumnToIndex[q.DeviceId][q.ServiceId] = map[string]int{}
		}
		for j, c := range q.Columns {
			deviceServiceColumnToIndex[q.DeviceId][q.ServiceId][c.Name] = j
		}
	}
	getIndexes := func(deviceId string, serviceId string, columnName string) (i int, j int, err error) {
		if _, ok := deviceServiceToIndex[deviceId]; !ok {
			return i, j, errors.New("device not found in timescale query")
		}
		if _, ok := deviceServiceToIndex[deviceId][serviceId]; !ok {
			return i, j, errors.New("service not found in timescale query")
		}
		i = deviceServiceToIndex[deviceId][serviceId]
		if _, ok := deviceServiceColumnToIndex[deviceId]; !ok {
			return i, j, errors.New("device not found in timescale query")
		}
		if _, ok := deviceServiceColumnToIndex[deviceId][serviceId]; !ok {
			return i, j, errors.New("service not found in timescale query")
		}
		if _, ok := deviceServiceColumnToIndex[deviceId][serviceId][columnName]; !ok {
			return i, j, errors.New("column name not found in timescale query")
		}
		j = deviceServiceColumnToIndex[deviceId][serviceId][columnName]
		j = j + 1 //ignore first value which is a timestamp
		return i, j, nil
	}
	for _, req := range request {
		i, j, err := getIndexes(req.Device.Id, req.Service.Id, req.ColumnName)
		if err != nil {
			return result, err
		}
		if len(response)+1 < i {
			return result, errors.New("timescale: request i index to large for response")
		}
		if len(response[i]) != 1 {
			result = append(result, messages.TimescaleResponse{
				Value: nil,
			})
			continue
		}
		if len(response[i][0])+1 < j {
			return result, errors.New("timescale: request j index to large for response")
		}
		result = append(result, messages.TimescaleResponse{
			Value: response[i][0][j],
		})
	}
	return result, nil
}

type TimescaleQuery = []QueriesRequestElement

type QueriesRequestElement struct {
	DeviceId  string
	ServiceId string
	Limit     int
	Columns   []QueriesRequestElementColumn
}

type QueriesRequestElementColumn struct {
	Name string
}

type TimescaleQueryResponse = [][][]interface{} //[query-index][0][column-index + 1]
