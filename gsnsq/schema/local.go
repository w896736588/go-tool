package schema

import (
	"sync"
	"time"

	"github.com/w896736588/go-tool/gstool"
)

type Local struct {
	TopicChannelMap map[string]string
	lock            *sync.RWMutex  //锁
	Log             *gstool.GsSlog //日志
}

// AllowMainIdTopicChannel 检查是否能够执行 只有一种情况下不允许执行
func (h Local) AllowMainIdTopicChannel(mainId, topic, channel string) bool {
	h.lock.Lock()
	defer h.lock.Unlock()
	//判断mainId是否绑定了topic+channel以及这些topic+channel是否在线
	boolCheckRefuse := false
	assignTopicChannelList := make([]string, 0)
	for _, mainIdVal := range h.TopicChannelMap {
		if mainIdVal == mainId {
			boolCheckRefuse = true //说明
			assignTopicChannelList = append(assignTopicChannelList, topic+`|`+channel)
		}
	}
	if !boolCheckRefuse {
		h.PushDebugLog(`对任务mainId:` + mainId + `不进行检测拦截，未找到绑定的topic+channel`)
		return true
	}
	//如果当前topic+channel在 那么允许执行
	if gstool.ArrayExistValue(&assignTopicChannelList, topic+`|`+channel) {
		return true
	}
	return false
}

func (h Local) PushErrorLog(msg string) {
	if h.Log == nil {
		gstool.FmtPrintlnLog(msg)
	} else {
		h.Log.Errof(msg)
	}
}

func (h Local) PushDebugLog(msg string) {
	if h.Log == nil {
		gstool.FmtPrintlnLog(msg)
	} else {
		h.Log.Debugf(msg)
	}
}

func (h Local) Init() {
	h.TopicChannelMap = make(map[string]string)
}

func (h Local) OnlineTopicChannel(topic, channel string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	//如果已经存在 那么清空
	if _, ok := h.TopicChannelMap[topic+`|`+channel]; ok {
		h.PushErrorLog(`重复注册topic_channel` + topic + `|` + channel)
	}
	h.TopicChannelMap[topic+`|`+channel] = `` //未被使用
	time.Sleep(time.Millisecond)
}

func (h Local) OfflineTopicChannel(topic, channel string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.TopicChannelMap, topic+`|`+channel)
}

func (h Local) ReleaseTopicChannel(topic, channel string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.TopicChannelMap[topic+`|`+channel] = ``
}

// BalanceTopicChannel 获取topicChannel 注意使用完成后一定要调用 OfflineTopicChannel进行释放
func (h Local) BalanceTopicChannel(mainId string, concurrenceNum int) []string {
	h.lock.Lock()
	defer h.lock.Unlock()
	emptyTopicChannelList := make([]string, 0) //目前空闲的
	otherTopicChannelList := make([]string, 0)
	returnTopicChannelList := make([]string, concurrenceNum) //分配的
	for topicChannel, registerMainId := range h.TopicChannelMap {
		if registerMainId == mainId {
			returnTopicChannelList = append(returnTopicChannelList, topicChannel)
		} else if registerMainId == `` {
			emptyTopicChannelList = append(emptyTopicChannelList, topicChannel)
		} else {
			otherTopicChannelList = append(otherTopicChannelList, topicChannel)
		}
	}
	//已经刚好找到 那么返回
	if len(returnTopicChannelList) == concurrenceNum {
		return returnTopicChannelList
	}
	//如果找到的比需求的多 那么释放
	if len(returnTopicChannelList) > concurrenceNum {
		deleteNum := len(returnTopicChannelList) - concurrenceNum
		deleteTopicChannelList := returnTopicChannelList[0:deleteNum]
		for _, deleteTopicChannel := range deleteTopicChannelList {
			h.TopicChannelMap[deleteTopicChannel] = ``
		}
		return returnTopicChannelList[deleteNum:]
	}
	//继续分配empty的
	for _, TopicChannel := range emptyTopicChannelList {
		returnTopicChannelList = append(returnTopicChannelList, TopicChannel)
		if len(returnTopicChannelList) == concurrenceNum {
			return returnTopicChannelList
		}
	}
	//还是没有 那么再分配任意的
	//TODO 优化为根据使用量进行分配
	for _, TopicChannel := range otherTopicChannelList {
		returnTopicChannelList = append(returnTopicChannelList, TopicChannel)
		if len(returnTopicChannelList) == concurrenceNum {
			return returnTopicChannelList
		}
	}
	return returnTopicChannelList
}

// GetAssignInfo 返回分配明细
func (h Local) GetAssignInfo() string {
	return gstool.JsonEncode(h.TopicChannelMap)
}
