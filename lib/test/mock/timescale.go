/*
 * Copyright 2020 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/lib/timescale"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"strconv"
	"strings"
	"time"
)

type TimescaleMockResponses = map[string]map[string]map[string]interface{}

type TimescaleMock struct {
	count     int
	responses TimescaleMockResponses
}

func GetTimescaleMockFactory(responses TimescaleMockResponses) func(ctx context.Context, config util.Config) (timescale.Interface, error) {
	return func(ctx context.Context, config util.Config) (timescale.Interface, error) {
		return &TimescaleMock{responses: responses}, nil
	}
}

func Timescale(ctx context.Context, config util.Config) (timescale.Interface, error) {
	return &TimescaleMock{}, nil
}

func (this *TimescaleMock) Query(token string, request []messages.TimescaleRequest, timeout time.Duration) (result []messages.TimescaleResponse, err error) {
	for _, r := range request {
		count := this.count
		this.count = this.count + 1
		dResp, ok := this.responses[r.Device.Id]
		if !ok {
			log.Println("DEBUG: unknown device", r.Device.Id)
			result = append(result, messages.TimescaleResponse{})
			continue
		}
		sResp, ok := dResp[r.Service.Id]
		if !ok {
			log.Println("DEBUG: unknown service", r.Service.Id)
			result = append(result, messages.TimescaleResponse{})
			continue
		}
		cResp, ok := sResp[r.ColumnName]
		if !ok {
			log.Println("DEBUG: unknown column", r.ColumnName)
			result = append(result, messages.TimescaleResponse{})
			continue
		}
		if str, ok := cResp.(string); ok {
			str = strings.ReplaceAll(str, "{{count}}", strconv.Itoa(count))
			cResp = str
		}
		t := time.Time{}.Format(time.RFC3339)
		result = append(result, messages.TimescaleResponse{
			Time:  &t,
			Value: cResp,
		})
	}
	return
}
