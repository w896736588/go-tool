package lib_tool

type ChanStruct struct {
	ChanMaxNum int
	Chan       chan interface{}
	CallFunc   func(interface{})
}

//创建一个并发任务
func ChanCreate(chanMaxNum, maxDoNum int, callFunc func(interface{})) ChanStruct {
	chanStruct := ChanStruct{
		ChanMaxNum: chanMaxNum,
		Chan:       make(chan interface{}, chanMaxNum),
		CallFunc:   callFunc,
	}
	for i := 0; i < maxDoNum; i++ {
		go chanStruct.Do()
	}
	return chanStruct
}

//加入消息
func (h *ChanStruct) Add(msg interface{}) {
	h.Chan <- msg
}

//消费
func (h *ChanStruct) Do() {
	for {
		select {
		case msg := <-h.Chan:
			h.CallFunc(msg)
		}
	}
}
