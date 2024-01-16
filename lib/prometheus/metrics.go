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
	"os"
	"slices"
	"time"
)

type Metrics struct {
	IncidentsCount              prometheus.Counter
	TasksReceivedCount          prometheus.Counter
	TasksCompletedCount         prometheus.Counter
	TaskCompleteLatencyMs       prometheus.Gauge
	TasksCompleteErrors         prometheus.Counter
	GetTasksErrors              prometheus.Counter
	GetShardsError              prometheus.Counter
	CommandRoundtripMs          prometheus.Gauge
	CommandRoundtripMsHistogram prometheus.Histogram
	CommandResponsesReceived    prometheus.Counter

	TaskMarshallingLatency              *prometheus.HistogramVec
	TaskLastEventValueRequestCountVec   *prometheus.CounterVec
	TaskCommandSendCountVec             *prometheus.CounterVec
	TaskReceivedCountVec                *prometheus.CounterVec
	TaskCommandResponseReceivedCountVec *prometheus.CounterVec
	TaskCompletedCountVec               *prometheus.CounterVec

	httphandler http.Handler
	ignoreUsers []string
}

var instanceId string

func getInstanceId() string {
	if instanceId == "" {
		var err error
		instanceId, err = os.Hostname()
		if err != nil {
			instanceId = ""
		}
	}
	return instanceId
}

func NewMetrics(prefix string, ignoreUsers []string) *Metrics {
	reg := prometheus.NewRegistry()

	m := &Metrics{
		ignoreUsers: ignoreUsers,
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
		CommandRoundtripMsHistogram: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    prefix + "_command_roundtrip_ms_histogram",
			Help:    "duration histogram of command roundtrips in ms",
			Buckets: []float64{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1500, 2000, 2500, 3000},
		}),
		CommandResponsesReceived: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prefix + "_command_responses",
			Help: "count of command responses received since startup",
		}),

		TaskMarshallingLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "external_task_worker_task_marshalling_latency",
			Help:    "histogram vec for latency of marshaller calls",
			Buckets: []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2000},
		}, []string{"instance_id", "user_id", "endpoint", "service_id", "function_id"}),
		TaskLastEventValueRequestCountVec: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "external_task_worker_task_last_event_value_request_count_vec",
			Help: "counter vec for last-event-value requests",
		}, []string{"instance_id", "user_id", "process_definition_id"}),
		TaskCommandSendCountVec: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "external_task_worker_task_command_send_count_vec",
			Help: "counter vec for task commands send",
		}, []string{"instance_id", "user_id", "process_definition_id"}),
		TaskReceivedCountVec: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "external_task_worker_task_received_count_vec",
			Help: "counter vec for received tasks",
		}, []string{"instance_id", "user_id", "process_definition_id"}),
		TaskCommandResponseReceivedCountVec: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "external_task_worker_task_command_response_received_count_vec",
			Help: "counter vec for received command responses",
		}, []string{"instance_id", "user_id", "process_definition_id"}),
		TaskCompletedCountVec: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "external_task_worker_task_completed_count_vec",
			Help: "counter vec for completed tasks",
		}, []string{"instance_id", "user_id", "process_definition_id"}),
	}

	reg.MustRegister(m.IncidentsCount)
	reg.MustRegister(m.TasksReceivedCount)
	reg.MustRegister(m.TasksCompletedCount)
	reg.MustRegister(m.TaskCompleteLatencyMs)
	reg.MustRegister(m.TasksCompleteErrors)
	reg.MustRegister(m.GetTasksErrors)
	reg.MustRegister(m.GetShardsError)
	reg.MustRegister(m.CommandRoundtripMs)
	reg.MustRegister(m.CommandRoundtripMsHistogram)
	reg.MustRegister(m.CommandResponsesReceived)

	reg.MustRegister(
		m.TaskMarshallingLatency,
		m.TaskLastEventValueRequestCountVec,
		m.TaskCommandSendCountVec,
		m.TaskReceivedCountVec,
		m.TaskCommandResponseReceivedCountVec,
		m.TaskCompletedCountVec,
	)

	return m
}

func (this *Metrics) ignore(userId string) bool {
	return slices.Contains(this.ignoreUsers, userId)
}

func (this *Metrics) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Printf("%v [%v] %v \n", request.RemoteAddr, request.Method, request.URL)
	this.httphandler.ServeHTTP(writer, request)
}

func (this *Metrics) LogTaskMarshallingLatency(endpoint string, userId string, serviceId string, functionId string, latency time.Duration) {
	if this.ignore(userId) {
		return
	}
	this.TaskMarshallingLatency.WithLabelValues(getInstanceId(), userId, endpoint, serviceId, functionId).Observe(float64(latency.Milliseconds()))
}

func (this *Metrics) LogTaskLastEventValueRequest(task messages.GroupTaskMetadataElement) {
	if this.ignore(task.Task.TenantId) {
		return
	}
	this.TaskLastEventValueRequestCountVec.WithLabelValues(getInstanceId(), task.Task.TenantId, task.Task.ProcessDefinitionId).Inc()
}

func (this *Metrics) LogTaskCommandSend(task messages.GroupTaskMetadataElement) {
	if this.ignore(task.Task.TenantId) {
		return
	}
	this.TaskCommandSendCountVec.WithLabelValues(getInstanceId(), task.Task.TenantId, task.Task.ProcessDefinitionId).Inc()
}

func (this *Metrics) LogTaskReceived(task messages.CamundaExternalTask) {
	if this.ignore(task.TenantId) {
		return
	}
	this.TaskReceivedCountVec.WithLabelValues(getInstanceId(), task.TenantId, task.ProcessDefinitionId).Inc()
}

func (this *Metrics) LogTaskCommandResponseReceived(task messages.TaskInfo) {
	if this.ignore(task.TenantId) {
		return
	}
	this.TaskCommandResponseReceivedCountVec.WithLabelValues(getInstanceId(), task.TenantId, task.ProcessDefinitionId).Inc()
}

func (this *Metrics) LogTaskCompleted(task messages.TaskInfo) {
	if this.ignore(task.TenantId) {
		return
	}
	this.TaskCompletedCountVec.WithLabelValues(getInstanceId(), task.TenantId, task.ProcessDefinitionId).Inc()
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
		value := float64(time.Since(time.Unix(0, smallestTraceUnixTimestamp)).Milliseconds())
		this.CommandRoundtripMs.Set(value)
		this.CommandRoundtripMsHistogram.Observe(value)
	}
}
