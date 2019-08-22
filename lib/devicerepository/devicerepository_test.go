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

package devicerepository

import (
	"encoding/json"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetDeviceInfo(t *testing.T) {
	drcloser, deviceRepoUrl := DeviceRepoMock()
	defer drcloser()
	authcloser, authUrl := AuthMock()
	defer authcloser()
	permcloser, permUrl := PermsearchMock()
	defer permcloser()
	config := &util.ConfigStruct{DeviceRepoUrl: deviceRepoUrl, AuthEndpoint: authUrl, PermissionsUrl: permUrl}

	device, service, err := Factory.Get(config).GetDeviceInfo("device1", "service1", "user1")
	if err != nil {
		t.Fatal(err)
	}
	if device.Id != "device1" || device.Name != "device1.name" {
		t.Fatal("unexpected device", device)
	}
	if service.Id != "service1" || service.Name != "service1.name" {
		t.Fatal("unexpected service", service)
	}
}

func DeviceRepoMock() (closer func(), url string) {
	handler := http.NewServeMux()
	handler.HandleFunc("/devices/device1", func(writer http.ResponseWriter, request *http.Request) {
		json.NewEncoder(writer).Encode(model.DeviceInstance{Id: "device1", Name: "device1.name", DeviceType: "dt1"})
	})
	handler.HandleFunc("/device-types/dt1", func(writer http.ResponseWriter, request *http.Request) {
		json.NewEncoder(writer).Encode(model.DeviceType{Id: "dt1", Name: "dt1.name", Services: []model.Service{{Id: "service1", Name: "service1.name"}}})
	})
	handler.HandleFunc("/services/service1", func(writer http.ResponseWriter, request *http.Request) {
		json.NewEncoder(writer).Encode(model.Service{Id: "service1", Name: "service1.name"})
	})
	s := httptest.NewServer(handler)
	return s.Close, s.URL
}

func AuthMock() (closer func(), url string) {
	handler := http.NewServeMux()
	handler.HandleFunc("/auth/realms/master/protocol/openid-connect/token", func(writer http.ResponseWriter, request *http.Request) {
		json.NewEncoder(writer).Encode(OpenidToken{ExpiresIn: 1000000000000, RefreshExpiresIn: 100000000000, TokenType: "Bearer", RequestTime: time.Now(), AccessToken: "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIzaUtabW9aUHpsMmRtQnBJdS1vSkY4ZVVUZHh4OUFIckVOcG5CcHM5SjYwIn0.eyJqdGkiOiJiOGUyNGZkNy1jNjJlLTRhNWQtOTQ4ZC1mZGI2ZWVkM2JmYzYiLCJleHAiOjE1MzA1MzIwMzIsIm5iZiI6MCwiaWF0IjoxNTMwNTI4NDMyLCJpc3MiOiJodHRwczovL2F1dGguc2VwbC5pbmZhaS5vcmcvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJkZDY5ZWEwZC1mNTUzLTQzMzYtODBmMy03ZjQ1NjdmODVjN2IiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiMjJlMGVjZjgtZjhhMS00NDQ1LWFmMjctNGQ1M2JmNWQxOGI5IiwiYXV0aF90aW1lIjoxNTMwNTI4NDIzLCJzZXNzaW9uX3N0YXRlIjoiMWQ3NWE5ODQtNzM1OS00MWJlLTgxYjktNzMyZDgyNzRjMjNlIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJhZG1pbiIsImRldmVsb3BlciIsInVtYV9hdXRob3JpemF0aW9uIiwidXNlciJdfSwicmVzb3VyY2VfYWNjZXNzIjp7Im1hc3Rlci1yZWFsbSI6eyJyb2xlcyI6WyJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsInZpZXctcmVhbG0iLCJtYW5hZ2UtaWRlbnRpdHktcHJvdmlkZXJzIiwiaW1wZXJzb25hdGlvbiIsImNyZWF0ZS1jbGllbnQiLCJtYW5hZ2UtdXNlcnMiLCJxdWVyeS1yZWFsbXMiLCJ2aWV3LWF1dGhvcml6YXRpb24iLCJxdWVyeS1jbGllbnRzIiwicXVlcnktdXNlcnMiLCJtYW5hZ2UtZXZlbnRzIiwibWFuYWdlLXJlYWxtIiwidmlldy1ldmVudHMiLCJ2aWV3LXVzZXJzIiwidmlldy1jbGllbnRzIiwibWFuYWdlLWF1dGhvcml6YXRpb24iLCJtYW5hZ2UtY2xpZW50cyIsInF1ZXJ5LWdyb3VwcyJdfSwiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwicm9sZXMiOlsidW1hX2F1dGhvcml6YXRpb24iLCJhZG1pbiIsImNyZWF0ZS1yZWFsbSIsImRldmVsb3BlciIsInVzZXIiLCJvZmZsaW5lX2FjY2VzcyJdLCJuYW1lIjoiZGYgZGZmZmYiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJzZXBsIiwiZ2l2ZW5fbmFtZSI6ImRmIiwiZmFtaWx5X25hbWUiOiJkZmZmZiIsImVtYWlsIjoic2VwbEBzZXBsLmRlIn0.eOwKV7vwRrWr8GlfCPFSq5WwR_p-_rSJURXCV1K7ClBY5jqKQkCsRL2V4YhkP1uS6ECeSxF7NNOLmElVLeFyAkvgSNOUkiuIWQpMTakNKynyRfH0SrdnPSTwK2V1s1i4VjoYdyZWXKNjeT2tUUX9eCyI5qOf_Dzcai5FhGCSUeKpV0ScUj5lKrn56aamlW9IdmbFJ4VwpQg2Y843Vc0TqpjK9n_uKwuRcQd9jkKHkbwWQ-wyJEbFWXHjQ6LnM84H0CQ2fgBqPPfpQDKjGSUNaCS-jtBcbsBAWQSICwol95BuOAqVFMucx56Wm-OyQOuoQ1jaLt2t-Uxtr-C9wKJWHQ"})
	})
	handler.HandleFunc("/auth/admin/realms/master/users/user1/role-mappings/realm", func(writer http.ResponseWriter, request *http.Request) {
		json.NewEncoder(writer).Encode([]RoleMapping{{Name: "admin"}})
	})
	s := httptest.NewServer(handler)
	return s.Close, s.URL
}

func PermsearchMock() (closer func(), url string) {
	handler := http.NewServeMux()
	handler.HandleFunc("/jwt/check/deviceinstance/device1/x/bool", func(writer http.ResponseWriter, request *http.Request) {
		json.NewEncoder(writer).Encode(true)
	})
	s := httptest.NewServer(handler)
	return s.Close, s.URL
}
