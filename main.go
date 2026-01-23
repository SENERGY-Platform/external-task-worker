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

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/SENERGY-Platform/api-docs-provider/lib/client"
	"github.com/SENERGY-Platform/external-task-worker/docs"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda"
	"github.com/SENERGY-Platform/external-task-worker/lib/com/comswitch"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
	"github.com/SENERGY-Platform/external-task-worker/lib/timescale"

	"syscall"

	"github.com/SENERGY-Platform/external-task-worker/lib"
	"github.com/SENERGY-Platform/external-task-worker/util"
)

func main() {
	configLocation := flag.String("config", "config.json", "configuration file")
	flag.Parse()

	config, err := util.LoadConfig(*configLocation)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		sig := <-shutdown
		config.GetLogger().Info("received shutdown signal", "signal", sig)
		cancel()
	}()

	err = lib.StartCacheInvalidator(ctx, config)
	if err != nil {
		config.GetLogger().Warn("unable to start cache invalidator", "error", err)
	}

	if config.ApiDocsProviderBaseUrl != "" && config.ApiDocsProviderBaseUrl != "-" {
		err = PublishAsyncApiDoc(config)
		if err != nil {
			config.GetLogger().Error("unable to publish async api docs", "error", err.Error())
		}
	}

	lib.Worker(ctx, config, comswitch.Factory, devicerepository.Factory, camunda.Factory, marshaller.Factory, timescale.Factory)

	config.GetLogger().Info("worker stopped")
}

func PublishAsyncApiDoc(conf util.Config) error {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	return client.New(http.DefaultClient, conf.ApiDocsProviderBaseUrl).AsyncapiPutDoc(ctx, "github_com_SENERGY-Platform_external-task-worker", docs.AsyncApiDoc)
}
