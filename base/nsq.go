package base

import "github.com/nsqio/go-nsq"

func PublishMsg(pubHost, pubPort string, msg, topic string) {
	config := nsq.NewConfig()
	var err error
	var producer *nsq.Producer
	if producer, err = nsq.NewProducer(pubHost+`:`+pubPort, config); err != nil {
		return
	}
	producer.SetLoggerLevel(nsq.LogLevelError)
	if err := producer.Publish(topic, []byte(msg)); err != nil {
		return
	}
	producer.Stop()
}
