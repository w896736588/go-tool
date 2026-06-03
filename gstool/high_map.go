package gstool

import "github.com/easierway/concurrent_map"

type HighMap struct {
	cm *concurrent_map.ConcurrentMap
}

// Set 存储
func (m *HighMap) Set(key interface{}, value interface{}) {
	m.cm.Set(concurrent_map.StrKey(key.(string)), value)
}

// Get 读取
func (m *HighMap) Get(key interface{}) (interface{}, bool) {
	return m.cm.Get(concurrent_map.StrKey(key.(string)))
}

// Del 移除
func (m *HighMap) Del(key interface{}) {
	m.cm.Del(concurrent_map.StrKey(key.(string)))
}

// HighMapCreate 创建高性能map
func HighMapCreate(shardNum int) *HighMap {
	conMap := concurrent_map.CreateConcurrentMap(shardNum)
	return &HighMap{conMap}
}
