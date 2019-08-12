package kafka

import (
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
)

type FactoryType struct{}

var Factory = FactoryType{}

func (FactoryType) NewConsumer(config util.Config, listener func(msg string) error) (consumer ConsumerInterface, err error) {
	consumer, err = NewConsumer(config, config.ResponseTopic, func(topic string, msg []byte) error {
		return listener(string(msg))
	}, func(err error, consumer *Consumer) {
		log.Println("FATAL ERROR: kafka", err)
		log.Fatal(err)
	})
	return
}

func (FactoryType) NewProducer(config util.Config) (ProducerInterface, error) {
	return InitProducer(config)
}
