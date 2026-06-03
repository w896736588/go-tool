package gstask

import (
	"context"
	"golang.org/x/time/rate"
)

type TaskLimiter struct {
	Trys        int    //尝试次数
	limiterType string // second 每秒钟  minute 每分钟
	limiter     *rate.Limiter
}

/**
令牌桶：
每秒钟允许多少次请求
每分钟允许多少次请求
*/

// TaskSecondLimiter 创建一个秒级限流 每秒钟最多执行trys次
func TaskSecondLimiter(trys int) *TaskLimiter {
	taskLimiter := TaskLimiter{Trys: trys}
	taskLimiter.limiter = rate.NewLimiter(rate.Limit(float64(trys)), trys)
	taskLimiter.limiterType = `second`
	return &taskLimiter
}

// TaskMinuteLimiter 创建一个分钟级限流 每分钟最多执行trys次
func TaskMinuteLimiter(trys int) *TaskLimiter {
	taskLimiter := TaskLimiter{Trys: trys}
	taskLimiter.limiter = rate.NewLimiter(rate.Limit(float64(trys)/60.0), trys)
	taskLimiter.limiterType = `minute`
	return &taskLimiter
}

// RunWaitN 执行等待
func (h *TaskLimiter) RunWaitN(msg any, fun func(msg any)) error {
	waitError := h.limiter.WaitN(context.Background(), 1)
	if waitError != nil {
		return waitError
	}
	fun(msg)
	return nil
}
