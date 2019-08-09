package kafka

import "github.com/SENERGY-Platform/external-task-worker/util"

type FactoryInterface interface {
	NewConsumer(config util.ConfigType, listener func(msg string) error) (consumer ConsumerInterface, err error)
	NewProducer(config util.ConfigType)(ProducerInterface, error)
}

type ConsumerInterface interface {
	Stop()
}

type ProducerInterface interface {
	Produce(topic string, message string)
	Close()
}
