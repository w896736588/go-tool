package p_common

type Call struct {
	AllGlobal             func() ([]map[string]any, error)                    //全局变量
	GetSshConfig          func(any) (map[string]any, error)                   //查询ssh配置
	GetRedisConfig        func(any) (map[string]any, error)                   //查询redis配置
	GetAllSshConfig       func() ([]map[string]any, error)                    //查询所有ssh配置
	QueryGroupInfo        func(map[string]any) (map[string]any, error)        //查询分组信息
	QueryAccountList      func(map[string]any) ([]map[string]any, error)      //查询账号列表
	QueryMysqlConfig      func(int) (map[string]any, error)                   //查询mysql配置
	QueryGlobalConfig     func(map[string]any) (map[string]any, error)        //查询全局配置
	CmdInfo               func(any) (map[string]any, error)                   //cmd流程明细
	CreateSmartLastRecord func(map[string]any) (int64, error)                 //创建最后的使用记录
	UpdateSmartLastRecord func(map[string]any, map[string]any) (int64, error) //更新最后的使用记录
}
