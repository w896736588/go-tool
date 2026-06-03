package gslock

import (
	"context"
	"errors"
	"time"

	"github.com/w896736588/go-tool/gsdb"
	"github.com/w896736588/go-tool/gstool"
)

func NewRedisLock(redis *gsdb.GsRedis, expired time.Duration) *RedisLock {
	return &RedisLock{
		redis:   redis,
		expired: expired,
	}
}

type RedisLock struct {
	redis   *gsdb.GsRedis
	expired time.Duration
}

// GetLock 获取锁
// 返回值：
// bool 是否拿到锁
// string 如果未拿到锁，锁当前的值
// error 异常
func (h *RedisLock) GetLock(key, lockVal string) (bool, string, error) {
	isLock, err := h.redis.Client.SetNX(context.Background(), key, lockVal, h.expired).Result()
	if err != nil {
		return false, ``, gstool.Error(`获取锁失败 %s`, err.Error())
	} else if isLock == false {
		lockVal, err = h.redis.Client.Get(context.Background(), key).Result()
		return false, lockVal, err
	} else {
		return true, lockVal, nil
	}
}

// GetWaitLock 获取锁
// maxTry 尝试次数
// wait 没拿到锁时等待时间
// breakFunc 中断检测，如果返回了非空字符串则中断并且返回到第二个string返回值上
// 返回值：
// bool 是否拿到锁
// string 中断时返回的string
// error 异常
func (h *RedisLock) GetWaitLock(key, lockVal string, maxTry int, wait time.Duration, breakFunc func() string) (bool, string, error) {
	for i := 0; i < maxTry; i++ {
		if breakFunc != nil {
			exist := breakFunc()
			if exist != `` {
				return false, exist, nil
			}
		}
		ok, err := h.redis.Client.SetNX(context.Background(), key, lockVal, h.expired).Result()
		if err != nil {
			time.Sleep(wait)
			continue
		}
		if ok {
			return true, ``, nil
		}
	}
	return false, ``, errors.New(`获取锁失败`)
}

// ReleaseLock 释放锁
func (h *RedisLock) ReleaseLock(key, lockValString string) error {
	if lockValString == `` {
		return gstool.Error(`释放的接待锁的值为空，请检查逻辑`)
	}
	if h.redis.Client.Get(context.Background(), key).Val() == lockValString {
		return h.redis.Client.Del(context.Background(), key).Err()
	}
	return nil
}
