package gstool

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// RunIsBinary 拿到运行类型 go run 执行还是./build/xxx执行 true表示为编译后运行
func RunIsBinary() bool {
	osArgs := os.Args
	if len(osArgs) > 0 {
		return strings.Index(os.Args[0], `/build`) != -1
	}
	return false
}

func RunCallStacks(maxDepth int, skipRuntime bool, skip int, sep string) string {
	var stack []string
	for i := skip; ; i++ { // 从 skip 开始而不是 0
		if maxDepth > 0 && len(stack) >= maxDepth {
			break
		}
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		if skipRuntime {
			fnName := fn.Name()
			if strings.HasPrefix(fnName, "runtime.") ||
				strings.HasPrefix(fnName, "testing.") ||
				strings.Contains(fnName, "GetCallStack") {
				continue
			}
		}
		_, filename := filepath.Split(file)
		funcName := filepath.Base(fn.Name())
		stack = append(stack, fmt.Sprintf("%s:%d(%s)", filename, line, funcName))
	}
	return strings.Join(stack, sep)
}
