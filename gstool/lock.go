package gstool

import "sync"

type LockMap struct {
	LockMap sync.Map
}

func NewLockMap() *LockMap {
	return &LockMap{}
}

func (h *LockMap) GetLock(id any) func() {
	lock, _ := h.LockMap.LoadOrStore(id, &sync.Mutex{})
	mu := lock.(*sync.Mutex)
	mu.Lock() // 上锁
	return func() { mu.Unlock() }
}
