package gsnsq

import (
	"errors"

	"github.com/nsqio/go-nsq"
)

// NewProducer 创建一个生产者
func NewProducer(config *NsqConfig) (*nsq.Producer, error) {
	return NewProducerMaxInFlight(config, 1000)
}

// NewProducerMaxInFlight 创建一个生产者并设置最高并发
func NewProducerMaxInFlight(config *NsqConfig, maxInFlight int) (*nsq.Producer, error) {
	if config.PubMsgHost == `` {
		return nil, errors.New(`Publication message address cannot be empty `)
	}
	producerConfig := nsq.NewConfig()
	producerConfig.MaxInFlight = maxInFlight
	producer, producerErr := nsq.NewProducer(config.PubMsgHost, producerConfig)
	if producerErr != nil {
		return nil, producerErr
	}
	producer.SetLoggerLevel(nsq.LogLevelError)
	return producer, nil
}
