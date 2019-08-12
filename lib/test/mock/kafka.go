package mock

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/kafka"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"sync"
)

var Kafka = &KafkaMock{}

type KafkaMock struct {
	mux       sync.Mutex
	Produced  map[string][]string
	listeners []func(msg string) error
}

func (this *KafkaMock) NewConsumer(config util.Config, listener func(msg string) error) (consumer kafka.ConsumerInterface, err error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.listeners = append(this.listeners, listener)
	return this, nil
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
	this.Produced[topic] = append(this.Produced[topic], message)
}

func (this *KafkaMock) Close() {}

func (this *KafkaMock) Stop() {}

func (this *KafkaMock) SendToConsumer(msg string) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	for _, l := range this.listeners {
		err := l(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *KafkaMock) GetProduced(topic string) []string {
	this.mux.Lock()
	defer this.mux.Unlock()
	return this.Produced[topic]
}
