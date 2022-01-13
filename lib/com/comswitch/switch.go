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

package comswitch

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/com"
	"github.com/SENERGY-Platform/external-task-worker/lib/com/httpcommand"
	"github.com/SENERGY-Platform/external-task-worker/lib/com/kafka"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"strings"
)

type FactoryType struct{}

var Factory = FactoryType{}

func (this FactoryType) NewConsumer(basectx context.Context, config util.Config, responseListener func(msg string) error, errorListener func(msg string) error) (err error) {
	ctx, cancel := context.WithCancel(basectx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	used := false
	if config.HttpCommandConsumerPort != "" && config.HttpCommandConsumerPort != "-" && !config.DisableHttpConsumer {
		used = true
		err = httpcommand.Factory.NewConsumer(ctx, config, responseListener, errorListener)
		if err != nil {
			return err
		}
	}
	if config.KafkaUrl != "" && config.KafkaUrl != "-" && !config.DisableKafkaConsumer {
		used = true
		err = kafka.Factory.NewConsumer(ctx, config, responseListener, errorListener)
		if err != nil {
			return err
		}
	}
	if !used {
		return errors.New("no response consumer set; at least one of the following config fields must be set: HttpCommandConsumerPort, KafkaUrl; consumers my be disabled by DisableHttpConsumer or DisableKafkaConsumer")
	}
	return nil
}

func (this FactoryType) NewProducer(basectx context.Context, config util.Config) (result com.ProducerInterface, err error) {
	ctx, cancel := context.WithCancel(basectx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	kafkaProducer, err := kafka.Factory.NewProducer(ctx, config)
	if err != nil {
		return result, err
	}
	restProducer, err := httpcommand.Factory.NewProducer(ctx, config)
	if err != nil {
		return result, err
	}
	return &Producer{kafkaProducer: kafkaProducer, restProducer: restProducer}, nil
}

type Producer struct {
	kafkaProducer com.ProducerInterface
	restProducer  com.ProducerInterface
}

func (this Producer) getChildProducer(topic string) com.ProducerInterface {
	if IsHttpHandler(topic) {
		return this.restProducer
	}
	return this.kafkaProducer
}

func (this *Producer) Produce(topic string, message string) (err error) {
	return this.getChildProducer(topic).Produce(topic, message)
}

func (this *Producer) ProduceWithKey(topic string, key string, message string) (err error) {
	return this.getChildProducer(topic).ProduceWithKey(topic, key, message)
}

func IsHttpHandler(handler string) bool {
	return strings.HasPrefix(handler, "http://") || strings.HasPrefix(handler, "https://")
}
