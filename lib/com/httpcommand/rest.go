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

package httpcommand

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/SENERGY-Platform/external-task-worker/lib/com"
	"github.com/SENERGY-Platform/external-task-worker/util"
)

type FactoryType struct{}

var Factory = FactoryType{}

func (this FactoryType) NewConsumer(ctx context.Context, config util.Config, responseListener func(msg string) error, errorListener func(msg string) error) (err error) {
	router := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var handler func(msg string) error
		if request.Method == http.MethodPost && strings.TrimPrefix(request.URL.Path, "/") == config.ErrorTopic {
			handler = errorListener
		} else if request.Method == http.MethodPost && strings.TrimPrefix(request.URL.Path, "/") == "responses" {
			handler = responseListener
		} else {
			http.Error(writer, "unknown endpoint", http.StatusNotFound)
			return
		}
		msg, err := io.ReadAll(request.Body)
		if err != nil {
			config.GetLogger().Error("unable to read http response consumer body", "error", err)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if config.HttpCommandConsumerSync {
			err = handler(string(msg))
			if err != nil {
				config.GetLogger().Error("unable to handle http response consumer message", "error", err)
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			go func() {
				err = handler(string(msg))
				if err != nil {
					config.GetLogger().Error("unable to handle http response consumer message", "error", err)
					return
				}
			}()
		}

		writer.WriteHeader(http.StatusOK)
	})
	corsHandler := NewCors(router)
	logger := NewLogger(corsHandler)
	server := &http.Server{Addr: ":" + config.HttpCommandConsumerPort, Handler: logger, WriteTimeout: 10 * time.Second, ReadTimeout: 10 * time.Second, ReadHeaderTimeout: 2 * time.Second}
	go func() {
		config.GetLogger().Debug("starting http response consumer", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				config.GetLogger().Error("FATAL: http response consumer server error", "error", err)
				log.Fatal(err)
			} else {
				config.GetLogger().Info("http response consumer server closed")
			}
		}
	}()
	go func() {
		<-ctx.Done()
		log.Println("http response consumer shutdown", server.Shutdown(context.Background()))
	}()
	return nil
}

func (this FactoryType) NewProducer(ctx context.Context, config util.Config) (com.ProducerInterface, error) {
	return &Producer{config: config}, nil
}
