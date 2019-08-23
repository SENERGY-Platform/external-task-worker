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
