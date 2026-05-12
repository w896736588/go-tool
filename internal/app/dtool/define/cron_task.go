package define

const (
	CronTaskTypeDailyReport = `daily_report`
	CronTaskTypeMemorySync  = `memory_sync`
	CronTaskTypeMainDBSync  = `main_db_sync`
)

// CronTaskRegistry 定义所有定时任务类型及其名称和执行函数注册信息。
var CronTaskRegistry = map[string]CronTaskDef{
	CronTaskTypeDailyReport: {Name: `AI 生成工作日报`},
	CronTaskTypeMemorySync:  {Name: `同步知识片段（兜底）`},
	CronTaskTypeMainDBSync:  {Name: `同步主库（兜底）`},
}

// CronTaskDef 描述一种定时任务类型。
type CronTaskDef struct {
	Name string
}
