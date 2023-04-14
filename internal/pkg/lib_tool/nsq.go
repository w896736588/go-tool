package lib_tool

import (
	"errors"
	"github.com/nsqio/go-nsq"
	"github.com/spf13/cast"
	"strconv"
	"time"
)

type NsqStruct struct {
	Topic        string
	Channel      string
	PConfig      NsqConfig
	CConfig      NsqConfig
	ConsumerList []*nsq.Consumer
	Producer     *nsq.Producer
	ProducerChan ChanStruct
}

type NsqConfig struct {
	Host string
	Port string
}

//初始化基本的topic
func NsqInit(topic string) *NsqStruct {
	nsqHandle := NsqStruct{
		Topic:        topic,
		ConsumerList: make([]*nsq.Consumer, 0),
	}
	return &nsqHandle
}

//初始化消费者
func (h *NsqStruct) CreateConsumer(channel string, num int, cConfig NsqConfig, callFunc func(string) bool) (*NsqStruct, error) {
	nsqHandle := NsqStruct{
		Channel:      channel,
		CConfig:      cConfig,
		ConsumerList: make([]*nsq.Consumer, 0),
	}
	for i := 0; i < num; i++ {
		nsqConsumer, err := nsqHandle.createConsumer(callFunc)
		if err != nil {
			return &nsqHandle, err
		}
		nsqHandle.ConsumerList = append(nsqHandle.ConsumerList, nsqConsumer)
	}
	return &nsqHandle, nil
}

//创建发布者
func (h *NsqStruct) CreateProducer(pConfig NsqConfig) error {
	h.PConfig = pConfig
	producerConfig := nsq.NewConfig()
	var err error
	if h.Producer, err = nsq.NewProducer(h.PConfig.Host+`:`+h.PConfig.Port, producerConfig); err != nil {
		return err
	}
	h.Producer.SetLoggerLevel(nsq.LogLevelError)
	return nil
}

//启用异步并发推送消息
func (h *NsqStruct) StartChanProducer(chanMaxNum, maxDoNum int) {
	h.ProducerChan = ChanCreate(chanMaxNum, maxDoNum, func(msg interface{}) {
		err := h.PublishMsg(cast.ToString(msg))
		if err != nil {
			FmtPrintlnLog(`StartChanProducer` + err.Error())
		}
	})
}

type noopNSQLogger struct{}

// Output allows us to implement the nsq.Logger interface
func (l *noopNSQLogger) Output(int, string) error {
	return nil
}

//消息的处理
type messageHandlerStruct struct {
	//回调函数 当返回false的时候重试
	backFunc func(string) bool
	//是否退出 当进程收到退出信号时置为true
	boolStop bool
}

//初始化消费者
func (h *NsqStruct) createConsumer(backFunc func(string) bool) (*nsq.Consumer, error) {
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(h.Topic, h.Channel, config)
	if err != nil {
		return nil, err
	}
	//最大允许向几台台NSQD服务器接受消息，默认是1，要特别注意
	consumer.ChangeMaxInFlight(1)
	//日志级别
	consumer.SetLoggerLevel(nsq.LogLevelDebug)
	consumer.SetLogger(&noopNSQLogger{}, nsq.LogLevelError)
	//消息接收handle
	messageHandler := &messageHandlerStruct{backFunc: backFunc, boolStop: false}
	//定义这个消费者可以起多少个处理消息的handler，这个数值必须比ChangeMaxInFlight小，这里其实时多并发个消费者模式，相当于设置消费者的数量
	consumer.AddHandler(
		messageHandler,
	)
	//连接至nsqlookupd,这里可以时集群
	mqHost := h.CConfig.Host + `:` + h.CConfig.Port
	nsqlds := []string{mqHost}
	if err := consumer.ConnectToNSQLookupds(nsqlds); err != nil {
		return nil, err
	}
	return consumer, nil
}

// HandleMessage 消费者逻辑
func (h *messageHandlerStruct) HandleMessage(m *nsq.Message) (err error) {

	if len(m.Body) == 0 {
		return nil
	}
	nsqMsg := string(m.Body)

	runRet := h.backFunc(nsqMsg)
	if runRet == true {
		return nil
	} else {
		//返回非true的时候消息重发
		return errors.New(strconv.Itoa(1)) //消息将会重发
	}
}

//停止消费者
func (h *NsqStruct) ConsumerShutDown() {
	for _, consumer := range h.ConsumerList {
		consumer.Stop()
	}
}

//停止生产者
func (h *NsqStruct) ProducerStop() {
	h.Producer.Stop()
}

//发布消息
func (h *NsqStruct) PublishMsg(msg string) error {
	if err := h.Producer.Publish(h.Topic, []byte(msg)); err != nil {
		return err
	}
	return nil
}

//延时发布消息
func (h *NsqStruct) PublishMsgDeffer(msg string, delay time.Duration) error {
	if err := h.Producer.DeferredPublish(h.Topic, delay, []byte(msg)); err != nil {
		return err
	}
	return nil
}
