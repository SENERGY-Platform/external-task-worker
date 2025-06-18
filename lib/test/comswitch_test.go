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
	"context"
	"github.com/SENERGY-Platform/external-task-worker/lib/com/comswitch"
	"github.com/SENERGY-Platform/external-task-worker/lib/com/kafka"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/docker"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestComswitch(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, zkIp, err := docker.Zookeeper(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	zookeeperUrl := zkIp + ":2181"

	//kafka
	kafkaUrl, err := docker.Kafka(ctx, wg, zookeeperUrl)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(2 * time.Second)

	err = kafka.InitTopic(kafkaUrl, nil, "test")
	if err != nil {
		t.Error(err)
		return
	}

	err = kafka.InitTopic(kafkaUrl, nil, "test2")
	if err != nil {
		t.Error(err)
		return
	}

	messages := []string{}

	apiPort, err := getFreePort()
	if err != nil {
		t.Error(err)
		return
	}

	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.KafkaUrl = kafkaUrl
	config.KafkaConsumerGroup = "test"
	config.ResponseTopic = "test"
	config.HttpCommandConsumerPort = strconv.Itoa(apiPort)
	config.Debug = true
	config.InitTopics = true

	err = comswitch.Factory.NewConsumer(ctx, config, func(msg string) error {
		messages = append(messages, msg)
		return nil
	}, func(msg string) error {
		return nil
	})

	producer, err := comswitch.Factory.NewProducer(ctx, config)
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("http produce 1")
	err = producer.Produce("http://localhost:"+config.HttpCommandConsumerPort+"/responses", "http_msg1")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("http produce 2")
	err = producer.Produce("http://localhost:"+config.HttpCommandConsumerPort+"/responses", "http_msg2")
	if err != nil {
		t.Error(err)
		return
	}

	log.Println("kafka produce 1")
	err = producer.Produce("test", "msg1")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("kafka produce 2")
	err = producer.Produce("test", "msg2")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("kafka produce 3")
	err = producer.Produce("test2", "msg3")
	if err != nil {
		t.Error(err)
		return
	}

	log.Println("produced")

	time.Sleep(20 * time.Second)

	if !reflect.DeepEqual(messages, []string{"http_msg1", "http_msg2", "msg1", "msg2"}) {
		t.Error(messages)
	}
}

func TestComswitchProduceWithKey(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, zkIp, err := docker.Zookeeper(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	zookeeperUrl := zkIp + ":2181"

	//kafka
	kafkaUrl, err := docker.Kafka(ctx, wg, zookeeperUrl)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(2 * time.Second)

	err = kafka.InitTopic(kafkaUrl, nil, "test")
	if err != nil {
		t.Error(err)
		return
	}

	err = kafka.InitTopic(kafkaUrl, nil, "test2")
	if err != nil {
		t.Error(err)
		return
	}

	messages := []string{}
	errMessages := []string{}

	apiPort, err := getFreePort()
	if err != nil {
		t.Error(err)
		return
	}

	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.KafkaUrl = kafkaUrl
	config.KafkaConsumerGroup = "test"
	config.ResponseTopic = "test"
	config.ErrorTopic = "err"
	config.HttpCommandConsumerPort = strconv.Itoa(apiPort)
	config.Debug = true
	config.InitTopics = true

	err = comswitch.Factory.NewConsumer(ctx, config, func(msg string) error {
		messages = append(messages, msg)
		return nil
	}, func(msg string) error {
		errMessages = append(errMessages, msg)
		return nil
	})

	producer, err := comswitch.Factory.NewProducer(ctx, config)
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("http produce 1")
	err = producer.ProduceWithKey("http://localhost:"+config.HttpCommandConsumerPort+"/responses", "key", "http_msg1")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("http produce 2")
	err = producer.ProduceWithKey("http://localhost:"+config.HttpCommandConsumerPort+"/responses", "key", "http_msg2")
	if err != nil {
		t.Error(err)
		return
	}

	log.Println("http produce err 1")
	err = producer.ProduceWithKey("http://localhost:"+config.HttpCommandConsumerPort+"/err", "key", "http_err_msg1")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("http produce err 2")
	err = producer.ProduceWithKey("http://localhost:"+config.HttpCommandConsumerPort+"/err", "key", "http_err_msg2")
	if err != nil {
		t.Error(err)
		return
	}

	log.Println("kafka produce 1")
	err = producer.ProduceWithKey("test", "key", "msg1")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("kafka produce 2")
	err = producer.ProduceWithKey("test", "key", "msg2")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("kafka produce 3")
	err = producer.ProduceWithKey("test2", "key", "msg3")
	if err != nil {
		t.Error(err)
		return
	}

	log.Println("kafka produce err 1")
	err = producer.ProduceWithKey("err", "key", "err_msg1")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("kafka produce err 2")
	err = producer.ProduceWithKey("err", "key", "err_msg2")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("kafka produce err 3")
	err = producer.ProduceWithKey("err2", "key", "err_msg3")
	if err != nil {
		t.Error(err)
		return
	}

	log.Println("produced")

	time.Sleep(20 * time.Second)

	if !reflect.DeepEqual(messages, []string{"http_msg1", "http_msg2", "msg1", "msg2"}) {
		t.Error(messages)
	}

	if !reflect.DeepEqual(errMessages, []string{"http_err_msg1", "http_err_msg2", "err_msg1", "err_msg2"}) {
		t.Error(errMessages)
	}
}
