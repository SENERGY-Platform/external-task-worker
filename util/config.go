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

package util

import (
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const OPTIMISTIC = "optimistic"
const PESSIMISTIC = "pessimistic"
const CAMUNDA_VARIABLES_INCIDENT = "incident"
const CAMUNDA_VARIABLES_PAYLOAD = "payload"
const CAMUNDA_VARIABLES_OVERWRITE = "overwrite"

type Config struct {
	PrometheusPort                  string
	ShardsDb                        string `config:"secret"`
	DeviceRepoUrl                   string
	CompletionStrategy              string
	OptimisticTaskCompletionTimeout int64
	CamundaLongPollTimeout          int64
	CamundaWorkerTasks              int64
	CamundaFetchLockDuration        int64
	CamundaTopic                    string
	CamundaTaskResultName           string
	KafkaUrl                        string
	KafkaConsumerGroup              string
	ResponseTopic                   string
	AuthExpirationTimeBuffer        float64
	AuthEndpoint                    string
	AuthClientId                    string `config:"secret"`
	AuthClientSecret                string `config:"secret"`
	JwtPrivateKey                   string `config:"secret"`
	JwtExpiration                   int64
	JwtIssuer                       string
	UseHttpIncidentProducer         bool
	IncidentApiUrl                  string
	KafkaIncidentTopic              string
	Debug                           bool
	MarshallerUrl                   string
	HealthCheckPort                 string
	GroupScheduler                  string

	HandleMissingLastValueTimeAsError bool //set to true for mgw implementations of last-value / TimescaleWrapperUrl
	TimescaleWrapperUrl               string

	HttpCommandConsumerPort string
	HttpCommandConsumerSync bool
	MetadataResponseTo      string
	DisableKafkaConsumer    bool
	DisableHttpConsumer     bool

	AsyncFlushFrequency string
	AsyncCompression    string
	SyncCompression     string
	Sync                bool
	SyncIdempotent      bool
	PartitionNum        int64
	ReplicationFactor   int64
	AsyncFlushMessages  int64

	KafkaConsumerMaxWait  string
	KafkaConsumerMinBytes int64
	KafkaConsumerMaxBytes int64

	SubResultExpirationInSeconds int32
	SubResultDatabaseUrls        []string
	MemcachedTimeout             string
	MemcachedMaxIdleConns        int64

	ResponseWorkerCount int64

	MetadataErrorTo string
	ErrorTopic      string

	CacheTimeout                    string
	CacheInvalidationAllKafkaTopics []string
	DeviceKafkaTopic                string
	DeviceGroupKafkaTopic           string

	KafkaTopicConfigs map[string][]kafka.ConfigEntry

	IgnoreUserMetrics []string

	ApiDocsProviderBaseUrl string

	InitTopics bool
}

func LoadConfig(location string) (config Config, err error) {
	file, error := os.Open(location)
	if error != nil {
		log.Println("error on config load: ", error)
		return config, error
	}
	error = json.NewDecoder(file).Decode(&config)
	if error != nil {
		log.Println("invalid config json: ", error)
		return config, error
	}
	handleEnvironmentVars(&config)
	if config.CompletionStrategy == OPTIMISTIC && config.GroupScheduler != PARALLEL {
		log.Println("WARNING: CompletionStrategy == optimistic && GroupScheduler != parallel --> set GroupScheduler to parallel")
		config.GroupScheduler = PARALLEL
	}
	return config, err
}

var camel = regexp.MustCompile("(^[^A-Z]*|[A-Z]*)([A-Z][^A-Z]+|$)")

func fieldNameToEnvName(s string) string {
	var a []string
	for _, sub := range camel.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}
	return strings.ToUpper(strings.Join(a, "_"))
}

// preparations for docker
func handleEnvironmentVars(config *Config) {
	configValue := reflect.Indirect(reflect.ValueOf(config))
	configType := configValue.Type()
	for index := 0; index < configType.NumField(); index++ {
		fieldName := configType.Field(index).Name
		fieldConfig := configType.Field(index).Tag.Get("config")
		envName := fieldNameToEnvName(fieldName)
		envValue := os.Getenv(envName)
		if envValue != "" {
			if !strings.Contains(fieldConfig, "secret") {
				fmt.Println("use environment variable: ", envName, " = ", envValue)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Int64 {
				i, _ := strconv.ParseInt(envValue, 10, 64)
				configValue.FieldByName(fieldName).SetInt(i)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Int32 {
				i, _ := strconv.ParseInt(envValue, 10, 32)
				configValue.FieldByName(fieldName).SetInt(i)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.String {
				configValue.FieldByName(fieldName).SetString(envValue)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Bool {
				b, _ := strconv.ParseBool(envValue)
				configValue.FieldByName(fieldName).SetBool(b)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Float64 {
				f, _ := strconv.ParseFloat(envValue, 64)
				configValue.FieldByName(fieldName).SetFloat(f)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Slice {
				val := []string{}
				for _, element := range strings.Split(envValue, ",") {
					val = append(val, strings.TrimSpace(element))
				}
				configValue.FieldByName(fieldName).Set(reflect.ValueOf(val))
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Map {
				value := map[string]string{}
				for _, element := range strings.Split(envValue, ",") {
					keyVal := strings.Split(element, ":")
					key := strings.TrimSpace(keyVal[0])
					val := strings.TrimSpace(keyVal[1])
					value[key] = val
				}
				configValue.FieldByName(fieldName).Set(reflect.ValueOf(value))
			}
		}
	}
}
