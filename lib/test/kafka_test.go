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

package test

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/kafka"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/docker"
	"github.com/ory/dockertest/v3"
	"log"
	"testing"
	"time"
)

func TestProducer_Produce(t *testing.T) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Error(err)
		return
	}

	closeZk, _, zkIp, err := docker.Zookeeper(pool)
	if err != nil {
		t.Error(err)
		return
	}
	defer closeZk()
	zookeeperUrl := zkIp + ":2181"

	//kafka
	closeKafka, err := docker.Kafka(pool, zookeeperUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer closeKafka()

	time.Sleep(2 * time.Second)

	err = kafka.InitTopic(zookeeperUrl, "test")
	if err != nil {
		t.Error(err)
		return
	}

	err = kafka.InitTopic(zookeeperUrl, "test2")
	if err != nil {
		t.Error(err)
		return
	}

	result := [][]byte{}

	consumer, err := kafka.NewConsumer(zookeeperUrl, "test", "test", func(topic string, msg []byte, t time.Time) error {
		result = append(result, msg)
		return nil
	}, func(err error, consumer *kafka.Consumer) {
		t.Error(err)
	})
	defer consumer.Stop()

	producer, err := kafka.PrepareProducer(zookeeperUrl, true, true)
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("produce 1")
	err = producer.Produce("test", "msg1")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("produce 2")
	err = producer.Produce("test", "msg2")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("produce 3")
	err = producer.Produce("test2", "msg3")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("produced")

	time.Sleep(20 * time.Second)

	if len(result) != 2 {
		t.Error(len(result))
	}

	if len(result) > 0 && string(result[0]) != "msg1" {
		t.Error(string(result[0]))
	}
}

func TestProducer_ProduceWithKey(t *testing.T) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Error(err)
		return
	}

	closeZk, _, zkIp, err := docker.Zookeeper(pool)
	if err != nil {
		t.Error(err)
		return
	}
	defer closeZk()
	zookeeperUrl := zkIp + ":2181"

	//kafka
	closeKafka, err := docker.Kafka(pool, zookeeperUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer closeKafka()

	time.Sleep(2 * time.Second)

	err = kafka.InitTopic(zookeeperUrl, "test")
	if err != nil {
		t.Error(err)
		return
	}

	err = kafka.InitTopic(zookeeperUrl, "test2")
	if err != nil {
		t.Error(err)
		return
	}

	result := [][]byte{}

	consumer, err := kafka.NewConsumer(zookeeperUrl, "test", "test", func(topic string, msg []byte, t time.Time) error {
		result = append(result, msg)
		return nil
	}, func(err error, consumer *kafka.Consumer) {
		t.Error(err)
	})
	defer consumer.Stop()

	producer, err := kafka.PrepareProducer(zookeeperUrl, true, true)
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("produce 1")
	err = producer.ProduceWithKey("test", "key", "msg1")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("produce 2")
	err = producer.ProduceWithKey("test", "key", "msg2")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("produce 3")
	err = producer.ProduceWithKey("test2", "key", "msg3")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("produced")

	time.Sleep(20 * time.Second)

	if len(result) != 2 {
		t.Error(len(result))
	}

	if len(result) > 0 && string(result[0]) != "msg1" {
		t.Error(string(result[0]))
	}
}
