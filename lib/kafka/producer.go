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
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"os"
	"runtime/debug"

	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kazoo-go"
)

type Producer struct {
	producer sarama.AsyncProducer
}

func InitProducer(config util.Config) (producer *Producer, err error) {
	if config.SaramaLog == "true" {
		sarama.Logger = log.New(os.Stderr, "[Sarama] ", log.LstdFlags)
	}

	producer = &Producer{}
	var kz *kazoo.Kazoo
	kz, err = kazoo.NewKazooFromConnectionString(config.ZookeeperUrl, nil)
	if err != nil {
		debug.PrintStack()
		return producer, err
	}
	broker, err := kz.BrokerList()
	kz.Close()

	if err != nil {
		debug.PrintStack()
		return producer, err
	}

	sarama_conf := sarama.NewConfig()
	sarama_conf.Version = sarama.V0_10_0_1
	producer.producer, err = sarama.NewAsyncProducer(broker, sarama_conf)
	if err != nil {
		debug.PrintStack()
		return producer, err
	}

	return producer, nil
}

func (this *Producer) Produce(topic string, message string) {
	log.Println("produce kafka msg: ", topic, message)
	this.producer.Input() <- &sarama.ProducerMessage{Topic: topic, Key: nil, Value: sarama.StringEncoder(message), Timestamp: util.TimeNow()}
}

func (this *Producer) Close() {
	this.producer.Close()
}
