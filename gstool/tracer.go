package gstool

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const maxDepth = 1000    //最多记录1000个方法 防止内存过大
const maxChildren = 1000 //最多Add 1000条  防止内存过大

/*
	流程开始
	tracer := helper.NewHelperTracer()
	end := tracer.Start("test start")
	defer func() {
		fmt.Println(tracer.GetString())
	}()
	defer end()

	方法顶部
	end := trace.Start(`A`)
	defer end()

	方法中
	trace.Add(`休眠1秒`)
*/

// 事件类型：0 普通事件  1 span
type event struct {
	msg  string
	loc  string // file:line
	kind int
}

type spanNode struct {
	event
	cost     time.Duration
	children []*spanNode // 关键：指针切片，避免值拷贝
}

type Tracer struct {
	mu          sync.Mutex
	root        *spanNode
	stack       []*spanNode // 未结束的 span 栈
	MaxDepth    int         //最多记录多少层 默认1000
	MaxChildren int         //最多每次记录多少个事件 默认100
}

// NewHelperTracer 创建根节点
func NewHelperTracer() *Tracer {
	return &Tracer{
		stack:       make([]*spanNode, 0),
		MaxDepth:    maxDepth,
		MaxChildren: maxChildren,
	}
}

// 工具：快速看一眼当前深度
func (h *Tracer) curDepth() int {
	return len(h.stack)
}

// Start 开始一个 span，返回 end 函数
func (h *Tracer) Start(format string, args ...any) (end func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// 1. 深度超限 → 空操作
	if h.curDepth() >= h.MaxDepth {
		return func() {} // 直接返回空 end
	}
	//2. 添加栈帧信息
	msg := fmt.Sprintf(format, args...)
	_, file, line, _ := runtime.Caller(1)
	loc := fmt.Sprintf("%s:%d", filepath.Base(file), line)
	now := time.Now()

	node := &spanNode{
		event: event{msg: msg, loc: loc, kind: 1},
	}

	if h.root == nil { // 第一个节点作为 root
		h.root = node
	} else { // 挂到当前栈顶 span 下
		parent := h.stack[len(h.stack)-1]
		parent.children = append(parent.children, node)
	}
	h.stack = append(h.stack, node)

	return func() { // 调用即结束本 span
		node.cost = time.Since(now)
		h.stack = h.stack[:len(h.stack)-1]
	}
}

// Add 记录一个普通事件
func (h *Tracer) Add(format string, args ...any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.root == nil || len(h.stack) == 0 { // 增加栈空检查
		return
	}
	// 1. 深度超限
	if h.curDepth() >= h.MaxDepth {
		return
	}
	// 2. 添加事件
	var msg string
	if len(args) == 0 {
		msg = format // 反引号字符串原样用
	} else {
		msg = fmt.Sprintf(format, args...)
	}
	_, file, line, _ := runtime.Caller(1)
	loc := fmt.Sprintf("%s:%d", filepath.Base(file), line)

	parent := h.stack[len(h.stack)-1]
	// 3. 父节点孩子数超限
	if len(parent.children) >= h.MaxChildren {
		return
	}
	parent.children = append(parent.children, &spanNode{
		event: event{msg: msg, loc: loc, kind: 0},
	})
}

// GetString 打印整棵树
func (h *Tracer) GetString() string {
	if h.root == nil {
		return ""
	}
	var sb strings.Builder
	var walk func(n *spanNode, depth int)
	walk = func(n *spanNode, depth int) {
		indent := strings.Repeat("  ", depth)
		if n.kind == 0 { // 事件
			//sb.WriteString(fmt.Sprintf("%s· %s  [%s]\n", indent, n.msg, n.loc))
			sb.WriteString(fmt.Sprintf("%s· %s\n", indent, n.msg))
		} else { // span
			//sb.WriteString(fmt.Sprintf("%s↳ %s  [%s %s] (%v)\n", indent, n.msg, n.loc, gstool.TimeNowUnixToString(`Y/m/d H:i:s`), n.cost))
			sb.WriteString(fmt.Sprintf("%s↳ %s  [%s,%v]\n", indent, n.msg, TimeNowUnixToString(`Y/m/d H:i:s`), n.cost))
		}
		for _, child := range n.children {
			walk(child, depth+1)
		}
	}
	walk(h.root, 0)
	return sb.String()
}
