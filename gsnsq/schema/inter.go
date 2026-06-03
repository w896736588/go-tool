package schema

type Schema interface {
	Init()
	OnlineTopicChannel(topic, channel string)                       //注册topic channel
	OfflineTopicChannel(topic, channel string)                      //离线topic channel
	ReleaseTopicChannel(topic, channel string)                      //释放topic channel
	BalanceTopicChannel(mainId string, concurrenceNum int) []string //获取mainId的topic channel
	AllowMainIdTopicChannel(mainId, topic, channel string) bool     //检查是否允许mainId在当前的topic和channel执行
	GetAssignInfo() string                                          //获取分配信息
	PushErrorLog(msg string)
	PushDebugLog(msg string)
}
