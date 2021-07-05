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
	"github.com/SENERGY-Platform/external-task-worker/lib/com"
	"github.com/SENERGY-Platform/external-task-worker/lib/com/kafka"
	"github.com/SENERGY-Platform/external-task-worker/lib/com/rest"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"strings"
)

type FactoryType struct{}

var Factory = FactoryType{}

func (this FactoryType) NewConsumer(basectx context.Context, config util.Config, listener func(msg string) error) (err error) {
	ctx, cancel := context.WithCancel(basectx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	err = rest.Factory.NewConsumer(ctx, config, listener)
	if err != nil {
		return err
	}
	err = kafka.Factory.NewConsumer(ctx, config, listener)
	if err != nil {
		return err
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
	restProducer, err := rest.Factory.NewProducer(ctx, config)
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
	if strings.HasPrefix("http://", topic) || strings.HasPrefix("https://", topic) {
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
