package gsnsq

import (
	"errors"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/w896736588/go-tool/gstool"
)

type NsqStruct struct {
	Topic        string
	Channel      string
	Config       NsqConfig
	ConsumerList []*nsq.Consumer
	Producer     *nsq.Producer
	//对内
	stopCallFunc func(string, string)
	wg           *sync.WaitGroup
}

// NsqConfig nsq的配置
type NsqConfig struct {
	LookUpHost string //lookup注册地址:lookup注册端口  一般是 lookup地址:4161
	PubMsgHost string //发布 消息地址 注意是nsqd host:4150 创建producer时已经内置了最高1000的发布缓存
}

// CreateConsumer 初始化消费者
func (h *NsqStruct) CreateConsumer(num int, callFunc func(string, uint16) bool) error {
	for i := 0; i < num; i++ {
		nsqConsumer, err := h.createConsumer(callFunc)
		if err != nil {
			return err
		}
		h.ConsumerList = append(h.ConsumerList, nsqConsumer)
	}
	return nil
}

// SetStopCallBack 设置停止消费者回调通知
func (h *NsqStruct) SetStopCallBack(callFunc func(string, string)) {
	h.stopCallFunc = callFunc
}

// CreateProducer 创建发布者
func (h *NsqStruct) CreateProducer() error {
	producer, producerErr := NewProducer(&h.Config)
	if producerErr != nil {
		return producerErr
	}
	h.Producer = producer
	return nil
}

type noopNSQLogger struct{}

// Output allows us to implement the nsq.GsSlog interface
func (l *noopNSQLogger) Output(calldepth int, s string) error {
	gstool.FmtPrintlnLog(`consumer error log ：%s %d %s`, gstool.DateCurrent(), calldepth, s)
	return nil
}

// 消息的处理
type messageHandlerStruct struct {
	backFunc func(string, uint16) bool
	wg       *sync.WaitGroup
}

// 初始化消费者
func (h *NsqStruct) createConsumer(backFunc func(string, uint16) bool) (*nsq.Consumer, error) {
	if h.wg == nil {
		h.wg = &sync.WaitGroup{}
	}
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(h.Topic, h.Channel, config)
	if err != nil {
		return nil, err
	}
	//同一个消费者同时处理中最大消息数 默认为1 最稳定
	consumer.ChangeMaxInFlight(1)
	//日志级别
	consumer.SetLoggerLevel(nsq.LogLevelError)
	consumer.SetLogger(&noopNSQLogger{}, nsq.LogLevelError)
	//消息接收handle
	messageHandler := &messageHandlerStruct{backFunc: backFunc, wg: h.wg}
	//定义这个消费者可以起多少个处理消息的handler，这个数值必须比ChangeMaxInFlight小，这里其实时多并发个消费者模式，相当于设置消费者的数量
	consumer.AddHandler(
		messageHandler,
	)
	//连接至nsqlookupd,这里可以时集群
	lookupHostList := []string{h.Config.LookUpHost}
	if lookupErr := consumer.ConnectToNSQLookupds(lookupHostList); lookupErr != nil {
		return nil, lookupErr
	}
	return consumer, nil
}

// HandleMessage 消费者逻辑
func (h *messageHandlerStruct) HandleMessage(m *nsq.Message) (err error) {
	if len(m.Body) == 0 {
		return nil
	}
	nsqMsg := string(m.Body)
	h.wg.Add(1)
	defer h.wg.Done()
	isOk := h.backFunc(nsqMsg, m.Attempts)
	if isOk {
		return nil
	} else {
		return errors.New(`retry`)
	}
}

// ConsumerShutDown 停止消费者 nsq自带优雅退出逻辑，最大等待每个消息handle最多30秒，过了会强制退出
func (h *NsqStruct) ConsumerShutDown() {
	for _, consumer := range h.ConsumerList {
		consumer.Stop()
	}
	h.wg.Wait()
	if h.stopCallFunc != nil {
		h.stopCallFunc(h.Topic, h.Channel)
	}
}

// ProducerStop 停止生产者
func (h *NsqStruct) ProducerStop() {
	h.Producer.Stop()
}

// PublishMsg 发布消息
func (h *NsqStruct) PublishMsg(msg string) error {
	if err := h.Producer.Publish(h.Topic, []byte(msg)); err != nil {
		return err
	}
	return nil
}

// PublishMsgDeffer 延时发布消息
func (h *NsqStruct) PublishMsgDeffer(msg string, delay time.Duration) error {
	if err := h.Producer.DeferredPublish(h.Topic, delay, []byte(msg)); err != nil {
		return err
	}
	return nil
}
