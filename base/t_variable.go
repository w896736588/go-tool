package base

import (
	"gitee.com/Sxiaobai/gs/gstool"
	"sync"
	"time"
)

type TVariable struct {
	TaskList map[string]string
	lock     sync.RWMutex
	Log      *gstool.GsSlog
}

func NewVariable() *TVariable {
	return &TVariable{
		TaskList: make(map[string]string),
		Log:      gstool.NewSlogDefault(Component.Env.LogPath, `variable`),
	}
}

func (h *TVariable) StopAll() {
	h.lock.Lock()
	defer h.lock.Unlock()
	for k, _ := range h.TaskList {
		h.TaskList[k] = "stop"
	}
	time.Sleep(1) //等待1秒 把其他任务的输出断开玩
}

func (h *TVariable) Add(id string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.TaskList[id] = "run"
}

func (h *TVariable) Del(id string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.TaskList, id)
}

func (h *TVariable) Get(id string) string {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.TaskList[id]
}
