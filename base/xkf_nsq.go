package base

import (
	"github.com/nsqio/go-nsq"
	log "github.com/sirupsen/logrus"
	"time"
)

// XkfDeferredPublishMsg 小客服测试环境发布消息
// @auth frog
// @date 2022-12-07 18:20:19
// @param msg
// @param delay
// @param topic
func XkfDeferredPublishMsg(msg string, delay time.Duration, topic string) {
	host := `121.40.109.241:4150`
	config := nsq.NewConfig()
	var err error
	var producer *nsq.Producer
	if producer, err = nsq.NewProducer(host, config); err != nil {
		log.Errorf(err.Error())
		return
	}
	if err := producer.DeferredPublish(topic, delay, []byte(msg)); err != nil {
		log.Errorf("PublishMsg error %#v", err.Error())
		return
	}
	log.Infof("消息发布 %s %s %#v", topic, msg, delay)
	producer.Stop()
}

// XkfPublishMsg 小客服测试环境发布消息
// @auth frog
// @date 2022-12-07 18:22:34
// @param msg
// @param topic
func XkfPublishMsg(msg string, topic string) {
	host := `121.40.109.241:4150`
	log.Debugf(`向nsq推入消息 topic ` + topic + ` host ` + host)
	config := nsq.NewConfig()
	var err error
	var producer *nsq.Producer
	if producer, err = nsq.NewProducer(host, config); err != nil {
		log.Errorf(err.Error())
		return
	}
	if err := producer.Publish(topic, []byte(msg)); err != nil {
		log.Errorf("PublishMsg error %#v", err.Error())
		return
	}
	log.Infof("消息发布 %s %s ", topic, msg)
	producer.Stop()
}
