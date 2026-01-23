/*
 * Copyright (c) 2023 InfAI (CC SES)
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

package prometheus

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/SENERGY-Platform/external-task-worker/util"
)

func Start(ctx context.Context, config util.Config) (metrics *Metrics, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()

	metrics = NewMetrics("external_task_worker_"+config.CamundaTopic, config.IgnoreUserMetrics)

	if config.PrometheusPort == "" || config.PrometheusPort == "-" {
		return metrics, nil
	}

	router := http.NewServeMux()

	router.Handle("/metrics", metrics)

	server := &http.Server{Addr: ":" + config.PrometheusPort, Handler: router}
	go func() {
		config.GetLogger().Info("prometheus api started")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			config.GetLogger().Error("FATAL: prometheus api error", "error", err)
			debug.PrintStack()
			log.Fatal("FATAL:", err)
		}
	}()
	go func() {
		<-ctx.Done()
		config.GetLogger().Info("prometheus api shutdown", "result", server.Shutdown(context.Background()))
	}()
	return
}
