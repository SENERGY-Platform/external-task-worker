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
	"net/http/httptest"

	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
	"github.com/SENERGY-Platform/marshaller/lib/api"
	"github.com/SENERGY-Platform/marshaller/lib/config"
	"github.com/SENERGY-Platform/marshaller/lib/configurables"
	"github.com/SENERGY-Platform/marshaller/lib/controller"
	marshaller_service "github.com/SENERGY-Platform/marshaller/lib/marshaller"
	marshaller_service_v2 "github.com/SENERGY-Platform/marshaller/lib/marshaller/v2"
	"github.com/SENERGY-Platform/marshaller/lib/tests/mocks"
)

var Marshaller = MarshallerService{}

// MarshallerService runs the marshaller service itself in the test process and answers with
// the client of this worker pointed at it. The worker therefore crosses the boundary it
// crosses in production, which is what makes the test trustworthy in two ways: the request
// travels as json, so nothing the worker hands over stays shared with the service — the
// marshalling writes the value it marshals into the content variables of the service it is
// given, and a shared service would make every message report the value of whichever task
// marshalled last — and the client has to reach the endpoint that serves the operation,
// which the shared interface alone does not prove.
type MarshallerService struct{}

func (this MarshallerService) New(ctx context.Context, url string) marshaller.Interface {
	conceptRepo, err := mocks.NewMockConceptRepo(ctx)
	if err != nil {
		panic(err)
	}
	conf := config.Config{}
	m := marshaller_service.New(mocks.Converter{}, conceptRepo, mocks.DeviceRepo)
	mV2 := marshaller_service_v2.New(conf, mocks.Converter{}, conceptRepo)
	ctrl := controller.New(conf, m, mV2, configurables.New(conceptRepo), mocks.DeviceRepo, nil)
	server := httptest.NewServer(api.GetRouter(conf, ctrl, nil))
	go func() {
		<-ctx.Done()
		server.Close()
	}()
	return marshaller.New(server.URL)
}
