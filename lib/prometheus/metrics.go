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
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"time"
)

type Metrics struct {
	IncidentsCount           prometheus.Counter
	TasksReceivedCount       prometheus.Counter
	TasksCompletedCount      prometheus.Counter
	TaskCompleteLatencyMs    prometheus.Gauge
	TasksCompleteErrors      prometheus.Counter
	GetTasksErrors           prometheus.Counter
	GetShardsError           prometheus.Counter
	CommandRoundtripMs       prometheus.Gauge
	CommandResponsesReceived prometheus.Counter

	httphandler http.Handler
}

func NewMetrics(prefix string) *Metrics {
	reg := prometheus.NewRegistry()
	m := &Metrics{
		httphandler: promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{
				Registry: reg,
			},
		),

		IncidentsCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prefix + "_incidents",
			Help: "count of incidents since startup",
		}),
		TasksReceivedCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prefix + "_tasks_received",
			Help: "count of tasks received since startup",
		}),
		TasksCompletedCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prefix + "_tasks_completed",
			Help: "count of tasks completed since startup",
		}),
		TaskCompleteLatencyMs: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "_task_complete_latency_ms",
			Help: "latency of device check in ms",
		}),
		TasksCompleteErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prefix + "_tasks_complete_errors",
			Help: "count of errors while completing tasks since startup",
		}),
		GetTasksErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prefix + "_get_tasks_errors",
			Help: "count of errors while loading tasks since startup",
		}),
		GetShardsError: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prefix + "_get_shards_errors",
			Help: "count of errors while loading shards since startup",
		}),
		CommandRoundtripMs: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "_command_roundtrip_ms",
			Help: "duration of a command roundtrip in ms",
		}),
		CommandResponsesReceived: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prefix + "_command_responses",
			Help: "count of command responses received since startup",
		}),
	}

	reg.MustRegister(m.IncidentsCount)
	reg.MustRegister(m.TasksReceivedCount)
	reg.MustRegister(m.TasksCompletedCount)
	reg.MustRegister(m.TaskCompleteLatencyMs)
	reg.MustRegister(m.TasksCompleteErrors)
	reg.MustRegister(m.GetTasksErrors)
	reg.MustRegister(m.GetShardsError)
	reg.MustRegister(m.CommandRoundtripMs)
	reg.MustRegister(m.CommandResponsesReceived)

	return m
}

func (this *Metrics) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Printf("%v [%v] %v \n", request.RemoteAddr, request.Method, request.URL)
	this.httphandler.ServeHTTP(writer, request)
}

func (this *Metrics) LogCamundaCompleteTask(latency time.Duration) {
	this.TasksCompletedCount.Inc()
	this.TaskCompleteLatencyMs.Set(float64(latency.Milliseconds()))
}

func (this *Metrics) LogCamundaCompleteTaskError() {
	this.TasksCompleteErrors.Inc()
}

func (this *Metrics) LogIncident() {
	this.IncidentsCount.Inc()
}

func (this *Metrics) LogCamundaLoadedTasks(count int) {
	if count > 0 {
		this.TasksReceivedCount.Add(float64(count))
	}
}

func (this *Metrics) LogCamundaGetShardsError() {
	this.GetShardsError.Inc()
}

func (this *Metrics) LogCamundaGetTasksError() {
	this.GetTasksErrors.Inc()
}

func (this *Metrics) HandleResponseTrace(trace []messages.Trace) {
	this.CommandResponsesReceived.Inc()
	if len(trace) == 0 {
		return
	}
	var smallestTraceUnixTimestamp int64 = -1
	for _, t := range trace {
		if smallestTraceUnixTimestamp < 0 || t.Timestamp < smallestTraceUnixTimestamp {
			smallestTraceUnixTimestamp = t.Timestamp
		}
	}
	if smallestTraceUnixTimestamp > 0 {
		this.CommandRoundtripMs.Set(float64(time.Since(time.Unix(0, smallestTraceUnixTimestamp)).Milliseconds()))
	}
}
