package schema

import "github.com/w896736588/go-tool/gsdb"

type Redis struct {
	RedisClient *gsdb.GsRedis
}

func (r Redis) AllowMainIdTopicChannel(mainId, topic, channel string) bool {
	//TODO implement me
	panic("implement me")
}

func (r Redis) CheckMainIdTopicChannel(mainId, topic, channel string) {
	//TODO implement me
	panic("implement me")
}

func (r Redis) Init() {
	//TODO implement me
	panic("implement me")
}

func (r Redis) OnlineTopicChannel(topic, channel string) {
	//TODO implement me
	panic("implement me")
}

func (r Redis) OfflineTopicChannel(topic, channel string) {
	//TODO implement me
	panic("implement me")
}

func (r Redis) ReleaseTopicChannel(topic, channel string) {
	//TODO implement me
	panic("implement me")
}

func (r Redis) BalanceTopicChannel(mainId string, concurrenceNum int) []string {
	//TODO implement me
	panic("implement me")
}

func (r Redis) GetAssignInfo() string {
	//TODO implement me
	panic("implement me")
}

func (r Redis) PushErrorLog(msg string) {
	//TODO implement me
	panic("implement me")
}

func (r Redis) PushDebugLog(msg string) {
	//TODO implement me
	panic("implement me")
}
