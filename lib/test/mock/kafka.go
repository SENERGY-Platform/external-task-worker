package mock

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/kafka"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"sync"
)

func NewKafkaMock() *KafkaMockFactory{
	return &KafkaMockFactory{
		Produced: map[string][]string{},
	}
}

type KafkaMockFactory struct {
	mux       sync.Mutex
	Produced  map[string][]string
	listeners []func(msg string) error
}

func (this *KafkaMockFactory) NewConsumer(config util.ConfigType, listener func(msg string) error) (consumer kafka.ConsumerInterface, err error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.listeners = append(this.listeners, listener)
	return this, nil
}

func (this *KafkaMockFactory) NewProducer(config util.ConfigType) (kafka.ProducerInterface, error) {
	return this, nil
}

func (this *KafkaMockFactory) Produce(topic string, message string) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.Produced[topic] = append(this.Produced[topic], message)
}

func (this *KafkaMockFactory) Close() {}

func (this *KafkaMockFactory) Stop() {}

func (this *KafkaMockFactory) SendToConsumer(msg string)(error){
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