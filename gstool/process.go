package gstool

import (
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsdefine"
)

// ProcessFindNewPidContainName 根据name找到最新的进程ID
func ProcessFindNewPidContainName(findName any) int32 {
	processList := ProcessList()
	ArrayMapFilterContainField(&processList, `name`, findName)
	ArrayMapSort(&processList, `create_time`, gsdefine.SortDesc)
	if len(processList) > 0 {
		return cast.ToInt32(processList[0][`pid`])
	}
	return 0
}

// ProcessFindNewPidByName 根据name找到最新的进程ID
func ProcessFindNewPidByName(findName any) int32 {
	processList := ProcessList()
	ArrayMapFilterField(&processList, `name`, findName)
	ArrayMapSort(&processList, `create_time`, gsdefine.SortDesc)
	if len(processList) > 0 {
		return cast.ToInt32(processList[0][`pid`])
	}
	return 0
}

// ProcessList 获取进程列表
func ProcessList() []map[string]any {
	processList, err := process.Processes()
	list := make([]map[string]any, 0)
	if err != nil {
		return list
	}
	for _, proc := range processList {
		name, _ := proc.Name()
		createTime, _ := proc.CreateTime()
		ppid, _ := proc.Ppid()
		exe, _ := proc.Exe()
		cmd, _ := proc.Cmdline()
		list = append(list, map[string]any{
			`name`:        name,
			`create_time`: createTime,
			`pid`:         proc.Pid,
			`ppid`:        ppid,
			`cmd`:         cmd,
			`exe`:         exe,
		})
	}
	return list
}

// ProcessFindNewPidByPPid 根据父进程获取最近创建的进程ID
func ProcessFindNewPidByPPid(ppid any) int32 {
	processList := ProcessList()
	ArrayMapFilterContainField(&processList, `ppid`, ppid)
	ArrayMapSort(&processList, `create_time`, gsdefine.SortDesc)
	if len(processList) > 0 {
		return cast.ToInt32(processList[0][`pid`])
	}
	return 0
}
