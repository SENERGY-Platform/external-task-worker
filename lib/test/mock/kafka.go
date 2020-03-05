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
	"github.com/SENERGY-Platform/external-task-worker/lib/kafka"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"sync"
)

var Kafka = &KafkaMock{}

type KafkaMock struct {
	mux       sync.Mutex
	Produced  map[string][]string
	listeners map[string][]func(msg string) error
}

func (this *KafkaMock) NewConsumer(config util.Config, listener func(msg string) error) (consumer kafka.ConsumerInterface, err error) {
	this.Subscribe(config.ResponseTopic, listener)
	return this, nil
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

func (this *KafkaMock) NewProducer(config util.Config) (kafka.ProducerInterface, error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.Produced = map[string][]string{}
	return this, nil
}

func (this *KafkaMock) Produce(topic string, message string) {
	this.mux.Lock()
	defer this.mux.Unlock()
	log.Println("Produce", topic, message)
	this.Produced[topic] = append(this.Produced[topic], message)
	for _, l := range this.listeners[topic] {
		log.Println(l(message))
	}
}

func (this *KafkaMock) ProduceWithKey(topic string, key string, message string) {
	this.mux.Lock()
	defer this.mux.Unlock()
	log.Println("Produce", topic, message)
	this.Produced[topic] = append(this.Produced[topic], message)
	for _, l := range this.listeners[topic] {
		log.Println(l(message))
	}
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
