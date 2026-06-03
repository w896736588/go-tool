package gsnsq

import (
	"errors"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/w896736588/go-tool/gsdb"
	"github.com/w896736588/go-tool/gsnsq/schema"
	"github.com/w896736588/go-tool/gstool"
)

/**
1. topic以及channel的分配由某一个ID来取模处理或者固定，比如管理员ID，应用ID等
2. 通过hash来存储每一个topic和channel的占用情况，任务完成后将会将分数置为0
3. 当改变了topic和channel的个数时，某些任务绑定的topic和channel可能会找不到 这时候将会允许执行
4. 适用于大任务，解决某一个业务大量的请求导致把其他任务全部堵塞问题
*/

type BalanceSchema int

const BalanceLocal BalanceSchema = 1 //本地模式
const BalanceRedis BalanceSchema = 2 //多服务器模式

// Balance nsq 空闲分配 按照空闲来分配任务 每个topic都会有n个channel
type Balance struct {
	Logger        *gstool.GsSlog //日志
	BalanceSchema BalanceSchema  //模式
	//对内
	topicList     []string      //topic列表
	channelList   []string      //channel列表
	nsqStructList []*NsqStruct  //nsq基础的构建
	schemaCache   schema.Schema //分配计算
}

// NewNsqBalance 创建一个nsq空闲分配
func NewNsqBalance(balanceSchema BalanceSchema, redisClient *gsdb.GsRedis, log *gstool.GsSlog) (*Balance, error) {
	var schemaCache schema.Schema
	if balanceSchema == BalanceLocal {
		schemaCache = schema.Local{
			Log: log,
		}
	} else {
		schemaCache = schema.Redis{
			RedisClient: redisClient,
		}
	}
	schemaCache.Init()
	return &Balance{
		nsqStructList: make([]*NsqStruct, 0),
		schemaCache:   schemaCache,
	}, nil
}

// SetTopicChannelList 设置topic和channel
func (h *Balance) SetTopicChannelList(topicList, channelList []string) {
	h.topicList = topicList
	h.channelList = channelList
}

// NsqBalanceStartConsumer 初始化均衡消费者
func (h *Balance) NsqBalanceStartConsumer(config NsqConfig, callFunc func(string, uint16) bool) error {
	if h.topicList == nil || h.channelList == nil {
		return errors.New(`not exist topic or channel list`)
	}
	for _, topic := range h.topicList {
		for _, channel := range h.channelList {
			nsqStruct := NsqStruct{
				Topic:        topic,
				Channel:      channel,
				Config:       config,
				ConsumerList: make([]*nsq.Consumer, 0),
			}
			//生产者
			if config.PubMsgHost != `` {
				producerError := nsqStruct.CreateProducer()
				if producerError != nil {
					return producerError
				}
			}
			//停止通知
			nsqStruct.SetStopCallBack(h.ConsumerStopNotify)
			//发布第一个消息
			publishErr := nsqStruct.Producer.Publish(topic, []byte(``))
			if publishErr != nil {
				return publishErr
			}
			//等待1秒钟
			time.Sleep(time.Second)
			//启动消费者
			nsqConsumer, err := nsqStruct.createConsumer(callFunc)
			if err != nil {
				return err
			}
			//加入到列表
			nsqStruct.ConsumerList = append(nsqStruct.ConsumerList, nsqConsumer)
			h.nsqStructList = append(h.nsqStructList, &nsqStruct)
			//注册上线
			h.schemaCache.OnlineTopicChannel(topic, channel)
		}
	}
	return nil
}

// ConsumerStopNotify 消费者停止通知
func (h *Balance) ConsumerStopNotify(topic, channel string) {
	h.schemaCache.OfflineTopicChannel(topic, channel)
}

// DistributionTopic 为消息分配topic
func (h *Balance) DistributionTopic(mainId string, concurrencyNum int) []string {
	if concurrencyNum < 1 {
		concurrencyNum = 1
	}
	return h.schemaCache.BalanceTopicChannel(mainId, concurrencyNum)
}

// ReleaseTopicChannel 总任务执行完，释放
func (h *Balance) ReleaseTopicChannel(topic, channel string) {
	h.schemaCache.ReleaseTopicChannel(topic, channel)
}

// PublishAllMsg 向所有消费者发布消息
func (h *Balance) PublishAllMsg(msg string) error {
	for _, nsqStruct := range h.nsqStructList {
		publishErr := nsqStruct.Producer.Publish(nsqStruct.Topic, []byte(msg))
		if publishErr != nil {
			return publishErr
		}
	}
	return nil
}

// PublishMsg 向某个topic消费者发布消息
func (h *Balance) PublishMsg(topic, msg string) error {
	for _, nsqStruct := range h.nsqStructList {
		if nsqStruct.Topic != topic {
			continue
		}
		publishErr := nsqStruct.Producer.Publish(nsqStruct.Topic, []byte(msg))
		if publishErr != nil {
			return publishErr
		}
	}
	return nil
}
