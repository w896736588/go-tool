package gstask

import (
	"errors"
	"sync"
	"time"
)

type CallbackFunc struct {
	Id      string
	Func    func() *Result
	Timeout time.Duration
}

type Task struct {
	taskFuncList []CallbackFunc
	isFinish     bool
	lock         sync.Mutex
}

type Result struct {
	Id     string //任务id
	Result any    //传任意内容
	State  int    //用来进行分辨的状态位
	Err    error  //1失败，0成功
}

func NewTask() *Task {
	return &Task{
		taskFuncList: make([]CallbackFunc, 0),
		isFinish:     false,
	}
}

func (h *Task) Add(callbacks ...CallbackFunc) {
	for _, callback := range callbacks {
		if callback.Timeout == 0 { //默认60秒
			callback.Timeout = 60 * time.Second
		}
		h.taskFuncList = append(h.taskFuncList, callback)
	}
}

// RunOne 执行任务时任意一个任务返回 则返回
func (h *Task) RunOne() *Result {
	wg := sync.WaitGroup{}
	wg.Add(1)
	result := &Result{}
	for _, callback := range h.taskFuncList {
		go func(cb CallbackFunc) {
			resultVal := h.runTask(cb)
			defer h.lock.Unlock()
			h.lock.Lock()
			if !h.isFinish {
				h.isFinish = true
				result = resultVal
				wg.Done()
			}
		}(callback)
	}
	wg.Wait()
	return result
}

// RunAll 运行所有任务
func (h *Task) RunAll() []*Result {
	wg := sync.WaitGroup{}
	wg.Add(len(h.taskFuncList))
	resultList := make([]*Result, 0)
	for _, callback := range h.taskFuncList {
		go func(cb CallbackFunc) {
			result := h.runTask(cb)
			resultList = append(resultList, result)
			wg.Done()
		}(callback)
	}
	wg.Wait()
	return resultList
}

func (h *Task) runTask(fun CallbackFunc) *Result {
	lock := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	isFinish := false
	var result = &Result{}
	//处理超时
	go func() {
		time.Sleep(fun.Timeout)
		defer lock.Unlock()
		lock.Lock()
		if !isFinish {
			isFinish = true
			result.Err = errors.New(`error：timeout`)
			result.Result = ``
			wg.Done()
		}
	}()
	//处理结果
	go func(cb CallbackFunc) {
		resultVal := cb.Func()
		defer h.lock.Unlock()
		h.lock.Lock()
		if !isFinish {
			isFinish = true
			result = resultVal
			wg.Done()
		}
	}(fun)
	wg.Wait()
	result.Id = fun.Id
	return result
}
