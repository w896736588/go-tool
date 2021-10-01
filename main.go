package main

import (
	"fmt"
	"redis_manager/api/gin"
	"redis_manager/base"
)

func main() {
	base.InitConfig()
	base.InitRedis()
	router := gin.InitRouter()
	router.Run(fmt.Sprintf(`:%s`, base.ConfigViper.GetString(`run.port`)))
}
