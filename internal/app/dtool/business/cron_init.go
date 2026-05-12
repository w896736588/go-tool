package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

// cronTaskFuncRegistry 将定时任务类型映射到对应的执行函数（本地任务在此注册，其他在 controller 中注册）。
var cronTaskFuncRegistry = map[string]func(){
	define.CronTaskTypeMemorySync: cronMemorySyncNow,
	define.CronTaskTypeMainDBSync: cronMainDBSyncNow,
}

// StartCronScheduler 从数据库读取所有定时任务配置并逐个启动调度器。
func StartCronScheduler() {
	seedDefaultCronTasks()
	list, err := common.DbMain.CronTaskList()
	if err != nil {
		gstool.FmtPrintlnLogTime(`定时任务列表读取失败 %s`, err.Error())
		return
	}
	for _, row := range list {
		taskType := cast.ToString(row[`type`])
		enabled := cast.ToInt(row[`enabled`]) == 1
		triggerTime := strings.TrimSpace(cast.ToString(row[`trigger_time`]))
		startCronSchedulerByType(taskType, enabled, triggerTime)
	}
}

// StopCronScheduler 停止所有定时调度器。
func StopCronScheduler() {
	for taskType, scheduler := range component.CronSchedulers {
		scheduler.Stop()
		delete(component.CronSchedulers, taskType)
	}
}

// startCronSchedulerByType 为指定类型创建或重启调度器。
func startCronSchedulerByType(taskType string, enabled bool, triggerTime string) {
	// 停止旧的
	if old, ok := component.CronSchedulers[taskType]; ok {
		old.Stop()
		delete(component.CronSchedulers, taskType)
	}
	if !enabled || triggerTime == `` {
		return
	}
	// 优先从 component 全局注册表查找（controller 注册），其次从本地注册表查找
	taskFunc, ok := component.CronTaskFuncRegistry[taskType]
	if !ok {
		taskFunc, ok = cronTaskFuncRegistry[taskType]
		if !ok {
			return
		}
	}
	scheduler := common.NewCronScheduler()
	scheduler.SetTaskFunc(taskFunc)
	scheduler.Configure(true, triggerTime)
	component.CronSchedulers[taskType] = scheduler
}

// seedDefaultCronTasks 确保所有注册的定时任务类型在数据库中有记录。
func seedDefaultCronTasks() {
	for taskType, def := range define.CronTaskRegistry {
		one, _ := common.DbMain.CronTaskByType(taskType)
		if cast.ToInt(one[`id`]) > 0 {
			continue
		}
		_ = common.DbMain.CronTaskSave(taskType, def.Name, 0, ``)
	}
}

// cronMemorySyncNow 执行知识片段兜底同步。
func cronMemorySyncNow() {
	if component.MemoryRuntime == nil {
		return
	}
	if err := component.MemoryRuntime.SyncNow(); err != nil {
		gstool.FmtPrintlnLogTime(`定时任务-知识片段同步失败 %s`, err.Error())
		return
	}
	_ = common.DbMain.CronTaskUpdateLastTriggerTime(define.CronTaskTypeMemorySync)
	gstool.FmtPrintlnLogTime(`定时任务-知识片段同步完成`)
}

// cronMainDBSyncNow 执行主库兜底同步。
func cronMainDBSyncNow() {
	if component.MainDBAutoSyncRuntime == nil {
		return
	}
	if err := component.MainDBAutoSyncRuntime.SyncNow(); err != nil {
		gstool.FmtPrintlnLogTime(`定时任务-主库同步失败 %s`, err.Error())
		return
	}
	_ = common.DbMain.CronTaskUpdateLastTriggerTime(define.CronTaskTypeMainDBSync)
	gstool.FmtPrintlnLogTime(`定时任务-主库同步完成`)
}
