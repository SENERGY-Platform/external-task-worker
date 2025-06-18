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

package kafka

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"github.com/SENERGY-Platform/external-task-worker/lib/com"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"strings"
	"time"
)

type FactoryType struct{}

var Factory = FactoryType{}

func (FactoryType) NewConsumer(ctx context.Context, config util.Config, responseListener func(msg string) error, errorListener func(msg string) error) (err error) {
	maxWait, err := time.ParseDuration(config.KafkaConsumerMaxWait)
	if err != nil {
		return errors.New("unable to parse KafkaConsumerMaxWait as duration: " + err.Error())
	}
	if config.ResponseTopic != "" && config.ResponseTopic != "-" {
		err = NewConsumer(ctx, ConsumerConfig{
			InitTopics:     config.InitTopics,
			KafkaUrl:       config.KafkaUrl,
			GroupId:        config.KafkaConsumerGroup,
			Topic:          config.ResponseTopic,
			MinBytes:       int(config.KafkaConsumerMinBytes),
			MaxBytes:       int(config.KafkaConsumerMaxBytes),
			MaxWait:        maxWait,
			TopicConfigMap: config.KafkaTopicConfigs,
		}, func(topic string, msg []byte, time time.Time) error {
			return responseListener(string(msg))
		}, func(err error) {
			if err != context.Canceled {
				log.Println("FATAL ERROR: kafka", err)
				log.Fatal(err)
			} else {
				log.Println("KAFKA CTX CANCELED", err)
			}
		})
		if err != nil {
			return err
		}
	}

	if config.ErrorTopic != "" && config.ErrorTopic != "-" && !(strings.HasPrefix(config.ErrorTopic, "http://") || strings.HasPrefix(config.ErrorTopic, "https://")) {
		err = NewConsumer(ctx, ConsumerConfig{
			InitTopics:     config.InitTopics,
			KafkaUrl:       config.KafkaUrl,
			GroupId:        config.KafkaConsumerGroup,
			Topic:          config.ErrorTopic,
			MinBytes:       int(config.KafkaConsumerMinBytes),
			MaxBytes:       int(config.KafkaConsumerMaxBytes),
			MaxWait:        maxWait,
			TopicConfigMap: config.KafkaTopicConfigs,
		}, func(topic string, msg []byte, time time.Time) error {
			return errorListener(string(msg))
		}, func(err error) {
			if err != context.Canceled {
				log.Println("FATAL ERROR: kafka", err)
				log.Fatal(err)
			} else {
				log.Println("KAFKA CTX CANCELED", err)
			}
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (FactoryType) NewProducer(ctx context.Context, config util.Config) (com.ProducerInterface, error) {
	flushFrequency, err := time.ParseDuration(config.AsyncFlushFrequency)
	if err != nil {
		return nil, err
	}
	return PrepareProducerWithConfig(ctx, config.KafkaUrl, ProducerConfig{
		InitTopics:          config.InitTopics,
		AsyncFlushFrequency: flushFrequency,
		AsyncCompression:    getKafkaCompression(config.AsyncCompression),
		SyncCompression:     getKafkaCompression(config.SyncCompression),
		Sync:                config.Sync,
		SyncIdempotent:      config.SyncIdempotent,
		PartitionNum:        int(config.PartitionNum),
		ReplicationFactor:   int(config.ReplicationFactor),
		AsyncFlushMessages:  int(config.AsyncFlushMessages),
		TopicConfigMap:      config.KafkaTopicConfigs,
	})
}

func getKafkaCompression(compression string) sarama.CompressionCodec {
	switch strings.ToLower(compression) {
	case "":
		return sarama.CompressionNone
	case "-":
		return sarama.CompressionNone
	case "none":
		return sarama.CompressionNone
	case "gzip":
		return sarama.CompressionGZIP
	case "snappy":
		return sarama.CompressionSnappy
	}
	log.Println("WARNING: unknown compression", compression, "fallback to none")
	return sarama.CompressionNone
}
