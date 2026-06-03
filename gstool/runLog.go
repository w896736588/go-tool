package gstool

import "fmt"

type RunLog struct {
	logs []string
}

func NewRunLog() *RunLog {
	return &RunLog{
		logs: make([]string, 0),
	}
}

func (h *RunLog) Set(msg string, args ...interface{}) {
	if len(args) == 0 {
		h.logs = append(h.logs, TimeNowUnixToString(`Y/m/d H:i:s`)+` `+msg)
	} else {
		h.logs = append(h.logs, TimeNowUnixToString(`Y/m/d H:i:s`)+` `+fmt.Sprintf(msg, args...))
	}
}

func (h *RunLog) Get() []string {
	return h.logs
}
