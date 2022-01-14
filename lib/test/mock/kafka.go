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

package mock

import (
	"context"
	"github.com/SENERGY-Platform/external-task-worker/lib/com"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"sync"
)

func CleanKafkaMock() {
	Kafka = &KafkaMock{}
}

var Kafka = &KafkaMock{}

type KafkaMock struct {
	mux       sync.Mutex
	Produced  map[string][]string
	listeners map[string][]func(msg string) error
}

func (this *KafkaMock) Log(logger *log.Logger) {

}

func (this *KafkaMock) NewConsumer(ctx context.Context, config util.Config, respListener func(msg string) error, errListener func(msg string) error) (err error) {
	this.Subscribe(config.ResponseTopic, respListener)
	this.Subscribe(config.ErrorTopic, errListener)
	return nil
}

func (this *KafkaMock) Subscribe(topic string, listener func(msg string) error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	if this.listeners == nil {
		this.listeners = map[string][]func(msg string) error{}
	}
	log.Println("Subscribe to", topic)
	this.listeners[topic] = append(this.listeners[topic], listener)
}

func (this *KafkaMock) NewProducer(ctx context.Context, config util.Config) (com.ProducerInterface, error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.Produced = map[string][]string{}
	return this, nil
}

func (this *KafkaMock) Produce(topic string, message string) error {
	log.Println("Produce", topic, message)
	this.setProduced(topic, message)
	for _, l := range this.getListeners(topic) {
		log.Println(l(message))
	}
	return nil
}

func (this *KafkaMock) setProduced(topic string, message string) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.Produced[topic] = append(this.Produced[topic], message)
}

func (this *KafkaMock) getListeners(topic string) []func(msg string) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	return this.listeners[topic]
}

func (this *KafkaMock) ProduceWithKey(topic string, key string, message string) error {
	return this.Produce(topic, message)
}

func (this *KafkaMock) Close() {}

func (this *KafkaMock) Stop() {}

func (this *KafkaMock) GetProduced(topic string) []string {
	this.mux.Lock()
	defer this.mux.Unlock()
	defer func() {
		this.Produced[topic] = []string{}
	}()
	return this.Produced[topic]
}

func (this *KafkaMock) New() *KafkaMock {
	return &KafkaMock{}
}
