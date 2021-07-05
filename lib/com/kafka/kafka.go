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
	"github.com/SENERGY-Platform/external-task-worker/lib/com"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"time"
)

type FactoryType struct{}

var Factory = FactoryType{}

func (FactoryType) NewConsumer(ctx context.Context, config util.Config, listener func(msg string) error) (err error) {
	return NewConsumer(ctx, config.KafkaUrl, config.KafkaConsumerGroup, config.ResponseTopic, func(topic string, msg []byte, time time.Time) error {
		return listener(string(msg))
	}, func(err error, consumer *Consumer) {
		log.Println("FATAL ERROR: kafka", err)
		log.Fatal(err)
	})
}

func (FactoryType) NewProducer(ctx context.Context, config util.Config) (com.ProducerInterface, error) {
	return PrepareProducer(ctx, config.KafkaUrl, true, false, config.Debug)
}
