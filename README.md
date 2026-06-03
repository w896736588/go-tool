# go-tool

[English](./README.en.md)

## 安装

```shell
go get -u github.com/w896736588/go-tool
```

## 一、mysql/pgsql/sqlite3

### 1. 创建连接

```go
//第二个参数true表示开启自动转换字段类型
//Ssh非必需
mysql := gsdb.NewMysql(&gsdb.MysqlConfig{
    Name:     "t",
    Host:     "localhost",
    Port:     3306,
    Username: "root",
    Password: "123456",
    Dbname:   "test",
	SshBridge: gsssh.NewSshBridge(gsssh.NewSsh(&gsssh.SshConfig{
        Host:     "11.11.11.11",
        Port:     "22",
        UserName: "xxx",
        Password: "xxxx",
    })),
}, true)
//开启debug hook 所有执行的sql都将会输出完整日志
//非必须
mysql.RegisterDebugHook(func(sql string, err error){
    //....
})
//设置自定义连接配置 非必须
mysql.SetOpenFunc(func(db *sql.DB) {
    db.SetConnMaxLifetime(time.Hour)
})
err := mysql.CreateConn()
```

### 2. 快捷操作

#### 说明

- 快捷操作设计的目的是为了将一些简单的sql使用链式来方便的书写，尤其是like  in这种，复杂的sql使用原生sql查询
- 当开启了autoTrans时条件和设置的值都会根据表结构自动转换为对应类型 mysql本身也支持
- 支持的操作符 > < >= <= <> like in not in between

#### 示例

- 更新所有满足条件的

```go
upNum, queryErr := mysql.QuickUpdate(`tbl_user`, map[string]any{
    `id`: [20,21], //会自动转为in
	`username`: []any{`like` , `%1`},
    `username1`: []any{`not in` , []any{1,2,`3`}},
    `username1`: []any{`in` , []any{1,2,3}},
    `create_time`: []any{`between` , []any{12312,23422}},
    `rawsql#1`: []any{`rawsql`, `role_id1 = ?`, []any{`39`}}, //自定义rawsql#1为不重复的字符串，其中的预编译的值不会被自动转换类型
}, map[string]any{
    `role_id1`: `234`,
    `nickname`: ``,
}).Limit(1).Exec()
```

- 查询100条

```go
list , queryErr := mysql.QuickQuery(`tbl_user` , `*` , map[string]any{
	`id` : []any{`>` , 1},
}).OffsetLimit(0 , 100).All()
//group by 查询
ma0, err := client.QuickQuery(`tbl_staff`, `count(user_id) as total,parent_user_id`, map[string]interface{}{
        `_id`: []any{`>`, 0},
    }).GroupBy(`parent_user_id`).All()
if err != nil {
    gstool.FmtPrintlnLogTime("查询失败:%v", err)
    return
}
gstool.FmtPrintlnLogTime(`查询结果:%s`, gstool.JsonFormat(ma0))
```

- pgsql的插入特殊处理 获取新增的id

```go
id , queryErr := mysql.QuickInsert(`tbl_user`, map[string]any{} , `id`)
```

- 使用事务 支持传递给多个数据库操作

```go
err := msql.CreateConn()
if err != nil {
    return
}
tx, err := msql.GetTx()
if err != nil {
    return
}
id, err := msql.QuickCreate(`tbl_user`, map[string]any{}{
    `admin_id`: 10000,
}).Exec(tx)
if err != nil {
    _ = tx.Rollback()
    return
}
info, err := msql.QueryBySql(`select * from tbl_user where id = ?`,id).One(tx)
if err != nil {
    _ = tx.Rollback()
    return
}
_ = tx.Commit()
```

- 快捷查询join等连接

```go
q := msql.QuickQuery(`tbl_test1 r`, `r.id`, map[string]any{}).
    Join(`left join tbl_test2 u on u.id = r.xx and u.xxx = ?`, `xxxxx`).
    Join(`....`)
ret, err := q.One()
gstool.FmtPrintlnLog(`%s`, q.GetSql())
```

- 将结果提取为切片

```go
ma4, err := client.QuickQuery(`tbl_staff`, `*`, map[string]any{}{
        `_id`: []any{`>`, 0},
    }).Limit(10).ToSlice(`user_id`)
if err != nil {
    gstool.FmtPrintlnLogTime("查询失败:%v", err)
    return
}
gstool.FmtPrintlnLogTime(`查询结果:%s`, gstool.JsonFormat(ma4.ToIntFilter()))
```

- 将结果提取为map

```go
ma3, err := client.QuickQuery(`tbl_staff`, `*`, map[string]any{}{
    `_id`: []any{`>`, 0},
}).Limit(10).ToMap(`user_id`, `parent_user_id`)
if err != nil {
    gstool.FmtPrintlnLogTime("查询失败:%v", err)
    return
}
gstool.FmtPrintlnLogTime(`查询结果:%s`, gstool.JsonFormat(ma3.ToStringInt()))
```

- 将结果按照某个字段分组

```go
ma2, err := client.QuickQuery(`tbl_staff`, `*`, map[string]any{}{
    `_id`: []any{`>`, 0},
}).Limit(10).ToGroupSlice(`parent_user_id`)
if err != nil {
    gstool.FmtPrintlnLogTime("查询失败:%v", err)
    return
}
gstool.FmtPrintlnLogTime(`查询结果:%s`, gstool.JsonFormat(ma2))
```

- 将结果转为map

```go
ma1, err := client.QuickQuery(`tbl_staff`, `*`, map[string]any{}{
    `_id`: []any{`>`, 0},
}).Limit(10).ToMapMap(`user_id`)
if err != nil {
    gstool.FmtPrintlnLogTime("查询失败:%v", err)
    return
}
gstool.FmtPrintlnLogTime(`查询结果:%s`, gstool.JsonFormat(ma1))
```

- 获取单个字段值

```go
ma0, err := client.QuickQuery(`tbl_staff`, `*`, map[string]any{}{
    `_id`: []any{`>`, 0},
}).Value(`user_id`)
if err != nil {
    gstool.FmtPrintlnLogTime("查询失败:%v", err)
    return
}
gstool.FmtPrintlnLogTime(`查询结果:%d`, ma0)
```

### 3. 获取实际执行的sql日志

- 获取单条执行的日志

```go
q := mysql.QuickSelect(`tbl_user`, []string{
	`id` : []any{`>` , 1}
}).OffsetLimit(0 , 100)
_,_ := q.All()
sql : q.GetSql() //获取实际执行的sql
```

- 注册全局日志

```go
client.RegisterDebugHook(func(sql string, err error) {
    gstool.FmtPrintlnLogTime(`sql %s`, sql)
    gstool.FmtPrintlnLogTime(`error %v`, err)
})
```

### 4. 自动转换类型说明

- 自动转换包括in not in 等各类操作的值，包括查询条件 更新的值
- 自动转换会查询表结构，如果遇到某个字段不存在时会重新更新表结构  如果更换了字段类型需要重新启动，否则可能导致插入类型错误，有可能执行失败
- 如果某个字段的类型变更，那么需要重新启动服务，自动转换可能会转换为错误类型
- rawsql类型的第二个参数，不会参与自动转换类型

### 5. 其他注意事项

- "date", "datetime", "timestamp"类型字段，需要传入string，不要直接传time.Time

## 二、http客户端

### 1. GET

```go
gshttp.Get(`http://xxxx/api`).Result()
```

### 2. POST

#### 提交数组的方式(application/x-www-form-urlencoded或multipart/form-data)

```go
//1.第一种方法 通过多次执行BodyMap()方法 可以给同一个key设置多个值 自动转为数组
gshttp.PostForm(`http://xxxx.api`).
	BodyMap(map[string]any{}).BodyMap(map[string]any{}).Request(5).Result()
//2.第二种方法 通过设置数组参数，将自动转为数组
gshttp.PostForm(`http://xxxx.api`).
	BodyMap(map[string]any{
	`params` : []string{`a`, `b`}
}).Request(5).Result()
```

#### application/json请求

```go
gshttp.PostJson(`http://xxxx.api`).
	BodyStr(`{"appid" : 1}`).Result()
```

#### application/x-www-form-urlencoded请求

```go
gshttp.PostForm(`http://xxxx.api`).
	BodyMap(map[string]any{}).Request(5).Result()
```

#### multipart/form-data方式提交 支持上传多个文件

```go
gshttp.PostMultiForm(`http://xxxx.api`).BodyMap(map[string]any{}).
	BodyFile(`file` , `本地地址` , `xxx.png`).
	BodyFile(`file` , `本地地址` , `xxx.png`).Request(5).Result()
```

#### 示例微信上传文件获取素材id

```go
func HttpWxPostFile(url, body string) (map[string]any,error) {
    data := make(map[string]any)
    deErr := gstool.JsonDecode(body, &data)
    if deErr != nil {
        return nil,deErr
    }
    ossUrl := cast.ToString(data[`oss_url`])
    if ossUrl == `` {
        return nil,errors.New(`oss_url不能为空`))
    }
    fileContent, contentErr := gstool.UrlGetContent(ossUrl)
    if contentErr != nil {
        return nil,contentErr
    }
    fileName := cast.ToString(data[`file_name`])
    if fileName == `` {
        return nil,errors.New(fmt.Sprintf(`file_name不能为空`))
    }
    ext := path.Ext(fileName)
    if ext == `` {
        return nil,errors.New(fmt.Sprintf(`文件后缀名获取失败`))
    }
    targetFileName := gstool.Md5(cast.ToString(fileContent)) + ext
    //临时存储的目录
    localFilePath := `/tmp/` + targetFileName
    crErr := gstool.FilePutContent(localFilePath, cast.ToString(fileContent))
    if crErr != nil {
        return nil,crErr
    }
    ret, err := gshttp.PostMultiForm(url).BodyFile(`media`, localFilePath, fileName).Request(20).ResultStr()
    if err != nil {
        return nil,err
    }
    dataM := make(map[string]any)
    dErr := gstool.JsonDecode(ret, &dataM)
    if dErr != nil {
        return nil,dErr
    }
    return dataM,nil
}
```

#### 允许非200的状态码

```go
gshttp.Get(`http://xxxx/api`).SetAllowHttpStatus(200 , 204).Result()
//keep-alive开启或者关闭 开启后 目标IP和端口相同时复用连接，不再需要三次握手 默认关闭
gshttp.Get(`http://xxxx/api`).OpenKeepAlive().Result()
```

#### 按流式接收

##### 按字符串作为分割

```go
fac := &stream.Byts{
    Byts: []byte("\n\n"),
        CallFunc: func(s string, err error) {
        h.StreamMsg(s, false)
    },
    FormatFunc: func(s []byte) []byte {
        if gstool.SContains(cast.ToString(s), []string{`忽略`}) { //这种内容不汇集到result结果中
            return []byte{} //返回的内容可以自己定义
        } else {
            return s //原样返回
        }
    },
}
result, err = gshttp.Get(`http://xxxx.api`).SetStreamFac(fac).Request(200).Result()
```

##### 按正则作为分割

```go
fac := &stream.Reges{
    Reges: `\x00{4}|[\x00-\x1F]`, //按照ascii分割
    CallFunc: func(s string, err error) {
        //分割得到的消息
    },
    FormatFunc: func(s []byte) []byte {
        if gstool.SContains(cast.ToString(s), []string{`忽略`}) { //这种内容不汇集到result结果中
            return []byte{} //返回的内容可以自己定义
        } else {
            return s //原样返回
        }
    },
}
result, err = gshttp.Get(`http://xxxx.api`).SetStreamFac(fac).Request(200).Result()
```

##### 按固定字节长度分割

```go
fac := &stream.BytsNum{
    Num: 255, //按照固定字节长度返回
    CallFunc: func(s string, err error) {
        //分割得到的消息
    },
    FormatFunc: func(s []byte) []byte {
        if gstool.SContains(cast.ToString(s), []string{`忽略`}) { //这种内容不汇集到result结果中
            return []byte{} //返回的内容可以自己定义
        } else {
            return s //原样返回
        }
    },
}
result, err = gshttp.Get(`http://xxxx.api`).SetStreamFac(fac).Request(200).Result()
```

## 三、并发执行通道

多个任务塞入执行队列中，可以选择执行所有或执行任一

```go
//执行所有
task := gstask.NewTask()
callBack := gstask.CallbackFunc{
    Func: func() *gstask.Result {
        //TODO 具体的执行的业务，需要返回gstask.Result类型
    },
    Timeout: 3 * time.Second, //超时时间
    Id:      0, //执行的业务的id 无论超时还是成功都会返回这个id
}
task.Add(callBack)
task.Add(callBack1) //继续往后添加
resultList := task.RunAll() //执行所有
result := task.RunOne() //执行，返回第一个执行完或者超时的结果
```

## 四、Nsq消费和发布消息

nsq不走协程并发，按照一个客户端同时只能处理一个消息处理

```go
nsqConfig := gsnsq.NsqConfig{
    LookUpHost: "127.0.0.1:4161", //消费地址
    PubMsgHost: "127.0.0.1:4150", //发布消息地址
}
nsq := gsnsq.NsqStruct{
    Topic:   "event",
    Channel: "event",
    Config:  nsqConfig,
}
//创建消费者 1表示创建1个消费者
_ = nsq.CreateConsumer(1, func(s string, att uint16) bool {
    //TODO 业务执行 s是消费者消息内容，att是重试次数，首次为1
    return true //返回true继续，返回false表示重新入队
})
//停止(所有消息消费完)
nsq.ConsumerShutDown()
//创建发布端 内置最高1000的并发发送缓冲区
_ = nsq.CreateProducer()
```

## 五、Redis

### 基础配置

```json
ssh := gsssh.SshConfig{
    Host:     `127.0.0.1`,
    Port:     `22`,
    UserName: `xxx`,
    Password: `xxx`,
}
redis := gsdb.GsRedis{
    RedisConfig: &gsdb.RedisConfig{
        Name:     "t",
        Host:     "127.0.0.1",
        Port:     6379,
        UserName: "xx",//高版本redis有账号名
        Password: "xxx",
    },
    SshBridge: gsssh.NewSshBridge(gsssh.NewSsh(&gsssh.SshConfig{
        Host:     "11.11.11.11",
        Port:     "22",
        UserName: "xxx",
        Password: "xxxx",
    })),
}
err := redis.CreateConn()
```

### redis锁（等待或非等待）

```go
lock := gslock.NewRedisLock(redisCli, time.Second*30)
// 一次性判断锁值并返回
// bool 是否拿到锁
// string 如果未拿到锁，锁当前的值
// error 异常
b, s, err := lock.GetLock(`锁的key值`, `锁的值`)

// 持续性判断锁值并返回
// maxTry 尝试次数
// wait 没拿到锁时下一次尝试拿锁的间隔时间
// breakFunc 中断拿锁方法 在每次拿锁之前会调用，如果返回了非空字符串则中断并且返回到第二个string返回值上；这个可以用来判断锁是否已经存在了，存在就中断最终返回
// 返回值：
// bool 是否拿到锁
// string 中断时返回的string，其他情况都返回空字符串
// error 异常
b, s, err := lock.GetWaitLock(`锁的key值`, `锁的值`, 10, time.Second, breakFunc)
```

### 缓存快速使用时构建

- func()可以换为查询数据库操作
- 遇到类型不匹配时将会返回传入类型的空值
- 泛型支持常用类型

#### hash存储map

```go
data, err := gsdb.RedisGetHashFromMap(client, `test1`, func() (map[string]map[string]string, error) {
    return map[string]map[string]string{
        `name`: {
            `name`: `xiaobai`,
        },
        `text`: {
            `text`: `hello world`,
        },
    }, nil
}, time.Hour)
```

```go
data, err := gsdb.RedisGetHashFromMap(client, `test1`, func() (map[string]string, error) {
    return map[string]string{
        `name`:  `xiaobai`,
        `text`:  `hello world`,
    }, nil
}, time.Hour)
```

#### string存储map

- func()可以换为查询数据库操作
- 遇到类型不匹配时将会返回传入类型的空值
- 泛型支持常用类型

```go
data, err := gsdb.RedisGetMapString(client, `test2`, func() (map[int]map[any]any, error) {
    return map[int]map[any]any{
        2: {
            "a": 1,
        },
    }, nil
}, time.Hour)
```

## 六、Ssh

### 一次性命令

```go
sshOnce := gsssh.NewSshOnce(gsssh.NewSsh(&gsssh.SshConfig{
    Host:     "11.11.11.11",
    Port:     "22",
    UserName: "xxx",
    Password: "xxxx",
}))
ret, err := sshOnce.RunCommandOnce(`ls -l`)
if err != nil {
    fmt.Println(err.Error())
    return
}
gstool.FmtPrintlnLogTime(`%s`, ret)
```

### 交互式

#### 交互式执行一个命令 接收所有返回 等待命令执行完

```go
sshTerminal := gsssh.NewSshTerminal(gsssh.NewSsh(&gsssh.SshConfig{
    Host:     "11.11.11.11",
    Port:     "22",
    UserName: "xxx",
    Password: "xxxx",
}))
//在这里设置回调会接收到所有的返回包括ssh链接成功的欢迎信息
sshTerminal.SetFuncStreamReceive(func(s string) {
    gstool.FmtPrintlnLogTime(`接收到内容 %s`, s)
})
//这里可以多次调用，不会在创建链接
//执行一个任务 查看当前文件夹情况 最多5秒钟，如果没有执行完则先返回 但是SetFuncStreamReceive会持续接收
ret, err := sshTerminal.RunCommandWait(`ls -l`, time.Second*5)
if err != nil {
    fmt.Println(err.Error())
    return
}
gstool.FmtPrintlnLogTime(`最终结果 %s`, ret)
```

#### 交互式执行一个命令 接收命令返回（不要ssh链接时的系统信息） 等待命令执行完

```go
sshTerminal := gsssh.NewSshTerminal(gsssh.NewSsh(&gsssh.SshConfig{
    Host:     "11.11.11.11",
    Port:     "22",
    UserName: "xxx",
    Password: "xxxx",
}))
_, err := sshTerminal.RunCommandWait(`pwd`, time.Second*10)
if err != nil {
    fmt.Println(err.Error())
    return
}
//这时候设置回调 那么就只会接收到后续命令的返回
sshTerminal.SetFuncStreamReceive(func(s string) {
    gstool.FmtPrintlnLogTime(`接收到内容 %s`, s)
})
//执行一个任务 查看当前文件夹情况 最多5秒钟，如果没有执行完则先返回 但是SetFuncStreamReceive会持续接收
//可以用于持续监听命令的返回 例如 tail -f /var/log/test/log 持续返回新内容
//如果当前交互式只用于一个命令 那么可以调用RunCommand 那么将不再阻塞返回ret，所有结果通过SetFuncStreamReceive接收
ret, err := sshTerminal.RunCommandWait(`ls -l`, time.Second*5)
if err != nil {
    fmt.Println(err.Error())
    return
}
gstool.FmtPrintlnLogTime(`最终结果 %s`, ret)
```

## 八、阿里云Oss

具体见gsali中的oss_client和oss_quick
