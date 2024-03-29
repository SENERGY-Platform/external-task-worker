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

package com

import (
	"context"
	"github.com/SENERGY-Platform/external-task-worker/util"
)

type FactoryInterface interface {
	NewConsumer(ctx context.Context, config util.Config, respoinseListener func(msg string) error, errorListener func(msg string) error) (err error)
	NewProducer(ctx context.Context, config util.Config) (ProducerInterface, error)
}

type ProducerInterface interface {
	Produce(topic string, message string) (err error)
	ProduceWithKey(topic string, key string, message string) (err error)
}
