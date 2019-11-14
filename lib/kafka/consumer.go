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
	"github.com/segmentio/kafka-go"
	"io"
	"io/ioutil"
	"log"
	"sync"

	"github.com/SENERGY-Platform/external-task-worker/util"

	"time"

	"github.com/wvanbergen/kazoo-go"
)

func NewConsumer(config util.Config, topic string, listener func(topic string, msg []byte) error, errorhandler func(err error, consumer *Consumer)) (consumer *Consumer, err error) {
	consumer = &Consumer{config: config, topic: topic, listener: listener, errorhandler: errorhandler}
	err = consumer.start()
	return
}

type Consumer struct {
	config       util.Config
	topic        string
	ctx          context.Context
	cancel       context.CancelFunc
	listener     func(topic string, msg []byte) error
	errorhandler func(err error, consumer *Consumer)
	mux          sync.Mutex
}

func (this *Consumer) Stop() {
	this.cancel()
}

func (this *Consumer) start() error {
	log.Println("DEBUG: consume topic: \"" + this.topic + "\"")
	this.ctx, this.cancel = context.WithCancel(context.Background())
	broker, err := GetBroker(this.config.ZookeeperUrl)
	if err != nil {
		log.Println("ERROR: unable to get broker list", err)
		return err
	}
	err = InitTopic(this.config.ZookeeperUrl, this.topic)
	if err != nil {
		log.Println("ERROR: unable to create topic", err)
		return err
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		CommitInterval: 0, //synchronous commits
		Brokers:        broker,
		GroupID:        this.config.KafkaConsumerGroup,
		Topic:          this.topic,
		MaxWait:        1 * time.Second,
		Logger:         log.New(ioutil.Discard, "", 0),
		ErrorLogger:    log.New(ioutil.Discard, "", 0),
	})
	go func() {
		for {
			select {
			case <-this.ctx.Done():
				log.Println("close kafka reader ", this.topic)
				return
			default:
				m, err := r.FetchMessage(this.ctx)
				if err == io.EOF || err == context.Canceled {
					log.Println("close consumer for topic ", this.topic)
					return
				}
				if err != nil {
					log.Println("ERROR: while consuming topic ", this.topic, err)
					this.errorhandler(err, this)
					return
				}
				if util.TimeNow().Sub(m.Time) > time.Duration(this.config.CamundaFetchLockDuration)*time.Millisecond {
					log.Println("DEBUG: kafka message older than CamundaFetchLockDuration --> ignore:", m.Time, string(m.Value))
					err = r.CommitMessages(this.ctx, m)
					if err != nil {
						log.Println("ERROR: while committing message ", this.topic, err, string(m.Value))
						this.errorhandler(err, this)
						return
					}
				} else {
					err = this.listener(m.Topic, m.Value)
					if err != nil {
						log.Println("ERROR: unable to handle message (no commit)", err, string(m.Value))
					} else {
						err = r.CommitMessages(this.ctx, m)
						if err != nil {
							log.Println("ERROR: while committing message ", this.topic, err, string(m.Value))
							this.errorhandler(err, this)
							return
						}
					}
				}
			}
		}
	}()
	return err
}

func (this *Consumer) Restart() {
	this.Stop()
	this.start()
}

func GetBroker(zk string) (brokers []string, err error) {
	return getBroker(zk)
}

func getBroker(zkUrl string) (brokers []string, err error) {
	zookeeper := kazoo.NewConfig()
	zookeeper.Logger = log.New(ioutil.Discard, "", 0)
	zk, chroot := kazoo.ParseConnectionString(zkUrl)
	zookeeper.Chroot = chroot
	if kz, err := kazoo.NewKazoo(zk, zookeeper); err != nil {
		return brokers, err
	} else {
		defer kz.Close()
		return kz.BrokerList()
	}
}

func GetKafkaController(zkUrl string) (controller string, err error) {
	zookeeper := kazoo.NewConfig()
	zookeeper.Logger = log.New(ioutil.Discard, "", 0)
	zk, chroot := kazoo.ParseConnectionString(zkUrl)
	zookeeper.Chroot = chroot
	kz, err := kazoo.NewKazoo(zk, zookeeper)
	if err != nil {
		return controller, err
	}
	controllerId, err := kz.Controller()
	if err != nil {
		return controller, err
	}
	brokers, err := kz.Brokers()
	kz.Close()
	if err != nil {
		return controller, err
	}
	return brokers[controllerId], err
}

func InitTopic(zkUrl string, topics ...string) (err error) {
	return InitTopicWithConfig(zkUrl, 1, 1, topics...)
}

func InitTopicWithConfig(zkUrl string, numPartitions int, replicationFactor int, topics ...string) (err error) {
	controller, err := GetKafkaController(zkUrl)
	if err != nil {
		log.Println("ERROR: unable to find controller", err)
		return err
	}
	if controller == "" {
		log.Println("ERROR: unable to find controller")
		return errors.New("unable to find controller")
	}
	initConn, err := kafka.Dial("tcp", controller)
	if err != nil {
		log.Println("ERROR: while init topic connection ", err)
		return err
	}
	defer initConn.Close()
	for _, topic := range topics {
		err = initConn.CreateTopics(kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
		})
		if err != nil {
			return
		}
	}
	return nil
}
