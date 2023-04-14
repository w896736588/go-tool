package lib_db

type RedisConfig struct {
	Name     string
	Host     string
	Password string
	PoolSize string
	Default  int
}

type MysqlConfig struct {
	Host              interface{} `json:"host"`
	Port              interface{} `json:"port"`
	Username          interface{} `json:"username"`
	Password          interface{} `json:"password"`
	Dbname            interface{} `json:"dbname"`
	PoolSize          interface{} `json:"poolsize"`
	MaxOpenConns      interface{} `json:"maxOpenConns"`
	MaxIdleConns      interface{} `json:"maxIdleConns"`
	MaxLifetimeSecond interface{} `json:"maxLifetimeSecond"` //链接重置时间 秒
}
