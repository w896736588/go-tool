# Installation
```go
go get -u github.com/w896736588/go-tool
```

# Update Plan
- Database operations support transactions OK
- In the quick operations of database operations, join series methods are supported OK
- Database operations support switching and conversion processing through connection OK
- Database operations support group by OK

## 1. mysql/pgsql/sqlite3
### 1. Create Connection

```go
// The second parameter true means enabling automatic field type conversion
// Ssh is optional
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
// Enable debug hook, all executed SQL will output complete logs
// Optional
mysql.RegisterDebugHook(func(sql string, err error){
    //....
})
// Set custom connection configuration, optional
mysql.SetOpenFunc(func(db *sql.DB) {
    db.SetConnMaxLifetime(time.Hour)
})
err := mysql.CreateConn()

```

### 2. Quick Operations
#### Description
- The purpose of quick operations is to write some simple SQL using chain methods to facilitate writing, especially for like and in operations. Complex SQL should use native SQL queries.
- When autoTrans is enabled, conditions and set values will be automatically converted to corresponding types according to the table structure. MySQL itself also supports this.
- Supported operators: > < >= <= <> like in not in between

#### Examples
- Update all records that meet the conditions
```go
upNum, queryErr := mysql.QuickUpdate(`tbl_user`, map[string]any{
    `id`: [20,21], // Will be automatically converted to in
	`username`: []any{`like` , `%1`},
    `username1`: []any{`not in` , []any{1,2,`3`}},
    `username1`: []any{`in` , []any{1,2,3}},
    `create_time`: []any{`between` , []any{12312,23422}},
    `rawsql#1`: []any{`rawsql`, `role_id1 = ?`, []any{`39`}}, // Custom rawsql#1 as non-repeating string, the precompiled values in it will not be automatically type converted
}, map[string]any{
    `role_id1`: `234`,
    `nickname`: ``,
}).Limit(1).Exec()
```

- Query 100 records
```go
list , queryErr := mysql.QuickQuery(`tbl_user` , `*` , map[string]any{
	`id` : []any{`>` , 1},
}).OffsetLimit(0 , 100).All()
// group by query
ma0, err := client.QuickQuery(`tbl_staff`, `count(user_id) as total,parent_user_id`, map[string]interface{}{
        `_id`: []any{`>`, 0},
    }).GroupBy(`parent_user_id`).All()
if err != nil {
    gstool.FmtPrintlnLogTime("Query failed:%v", err)
    return
}
gstool.FmtPrintlnLogTime(`Query result:%s`, gstool.JsonFormat(ma0))
```

- Special handling for pgsql insertion - Get the newly added id
```go
id , queryErr := mysql.QuickInsert(`tbl_user`, map[string]any{} , `id`)
```

- Using transactions - Support passing to multiple database operations
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

- Quick query joins and other connections
```go
q := msql.QuickQuery(`tbl_test1 r`, `r.id`, map[string]any{}).
    Join(`left join tbl_test2 u on u.id = r.xx and u.xxx = ?`, `xxxxx`).
    Join(`....`)
ret, err := q.One()
gstool.FmtPrintlnLog(`%s`, q.GetSql())
```

- Extract results as slices
```go
ma4, err := client.QuickQuery(`tbl_staff`, `*`, map[string]any{}{
        `_id`: []any{`>`, 0},
    }).Limit(10).ToSlice(`user_id`)
if err != nil {
    gstool.FmtPrintlnLogTime("Query failed:%v", err)
    return
}
gstool.FmtPrintlnLogTime(`Query result:%s`, gstool.JsonFormat(ma4.ToIntFilter()))
```

- Extract results as map
```go
ma3, err := client.QuickQuery(`tbl_staff`, `*`, map[string]any{}{
    `_id`: []any{`>`, 0},
}).Limit(10).ToMap(`user_id`, `parent_user_id`)
if err != nil {
    gstool.FmtPrintlnLogTime("Query failed:%v", err)
    return
}
gstool.FmtPrintlnLogTime(`Query result:%s`, gstool.JsonFormat(ma3.ToStringInt()))
```

- Group results by a specific field
```go
ma2, err := client.QuickQuery(`tbl_staff`, `*`, map[string]any{}{
    `_id`: []any{`>`, 0},
}).Limit(10).ToGroupSlice(`parent_user_id`)
if err != nil {
    gstool.FmtPrintlnLogTime("Query failed:%v", err)
    return
}
gstool.FmtPrintlnLogTime(`Query result:%s`, gstool.JsonFormat(ma2))
```

- Convert results to map
```go
ma1, err := client.QuickQuery(`tbl_staff`, `*`, map[string]any{}{
    `_id`: []any{`>`, 0},
}).Limit(10).ToMapMap(`user_id`)
if err != nil {
    gstool.FmtPrintlnLogTime("Query failed:%v", err)
    return
}
gstool.FmtPrintlnLogTime(`Query result:%s`, gstool.JsonFormat(ma1))
```

- Get a single field value
```go
ma0, err := client.QuickQuery(`tbl_staff`, `*`, map[string]any{}{
    `_id`: []any{`>`, 0},
}).Value(`user_id`)
if err != nil {
    gstool.FmtPrintlnLogTime("Query failed:%v", err)
    return
}
gstool.FmtPrintlnLogTime(`Query result:%d`, ma0)
```

### 3. Get actual executed SQL logs
- Get logs for a single execution
```go
q := mysql.QuickSelect(`tbl_user`, []string{
	`id` : []any{`>` , 1}
}).OffsetLimit(0 , 100)
_,_ := q.All()
sql : q.GetSql() // Get the actually executed SQL
```
- Register global log
```go
client.RegisterDebugHook(func(sql string, err error) {
    gstool.FmtPrintlnLogTime(`sql %s`, sql)
    gstool.FmtPrintlnLogTime(`error %v`, err)
})
```

### 4. Automatic Type Conversion Description
- Automatic conversion includes values for operations like in, not in, etc., including query conditions and update values
- Automatic conversion will query table structure. If a field doesn't exist, it will re-update the table structure. If field types are changed, service needs to be restarted, otherwise it may cause insertion type errors, possibly leading to execution failure
- If a field's type is changed, the service needs to be restarted as automatic conversion may convert to wrong type
- The second parameter of rawsql type will not participate in automatic type conversion

### 5. Other Notes
- For "date", "datetime", "timestamp" type fields, you need to pass in string, don't directly pass time.Time

## 2. HTTP Client
### 1. GET
```go
gshttp.Get(`http://xxxx/api`).Result()
```

### 2. POST
#### Submitting arrays (application/x-www-form-urlencoded or multipart/form-data)
```go
// 1. First method: By calling BodyMap() multiple times, you can set multiple values for the same key and automatically convert to an array
gshttp.PostForm(`http://xxxx.api`).
	BodyMap(map[string]any{}).BodyMap(map[string]any{}).Request(5).Result()
// 2. Second method: By setting array parameters, it will automatically convert to an array
gshttp.PostForm(`http://xxxx.api`).
	BodyMap(map[string]any{
	`params` : []string{`a`, `b`}   
}).Request(5).Result()
```
#### application/json request
```go
gshttp.PostJson(`http://xxxx.api`).
	BodyStr(`{"appid" : 1}`).Result()
```
#### application/x-www-form-urlencoded request
```go 
gshttp.PostForm(`http://xxxx.api`).
	BodyMap(map[string]any{}).Request(5).Result()
```
#### multipart/form-data submission - Supports uploading multiple files
```go
gshttp.PostMultiForm(`http://xxxx.api`).BodyMap(map[string]any{}).
	BodyFile(`file` , `local path` , `xxx.png`).
	BodyFile(`file` , `local path` , `xxx.png`).Request(5).Result()
```

#### Example: Upload file to WeChat to get material ID
```go
func HttpWxPostFile(url, body string) (map[string]any,error) {
    data := make(map[string]any)
    deErr := gstool.JsonDecode(body, &data)
    if deErr != nil {
        return nil,deErr
    }
    ossUrl := cast.ToString(data[`oss_url`])
    if ossUrl == `` {
        return nil,errors.New(`oss_url cannot be empty`))
    }
    fileContent, contentErr := gstool.UrlGetContent(ossUrl)
    if contentErr != nil {
        return nil,contentErr
    }
    fileName := cast.ToString(data[`file_name`])
    if fileName == `` {
        return nil,errors.New(fmt.Sprintf(`file_name cannot be empty`))
    }
    ext := path.Ext(fileName)
    if ext == `` {
        return nil,errors.New(fmt.Sprintf(`file extension name acquisition failed`))
    }
    targetFileName := gstool.Md5(cast.ToString(fileContent)) + ext
    // Temporary storage directory
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
#### Allow non-200 status codes
```go
gshttp.Get(`http://xxxx/api`).SetAllowHttpStatus(200 , 204).Result()
// Turn keep-alive on or off. When turned on, when the target IP and port are the same, connections are reused, no need for three-way handshake. Default is off
gshttp.Get(`http://xxxx/api`).OpenKeepAlive().Result()
```


#### Receive as stream
##### Split by string
```go
fac := &stream.Byts{
    Byts: []byte("\n\n"),
        CallFunc: func(s string, err error) {
        h.StreamMsg(s, false)
    },
    FormatFunc: func(s []byte) []byte {
        if gstool.SContains(cast.ToString(s), []string{`ignore`}) { // This content is not aggregated to the result
            return []byte{} // Return content can be customized
        } else {
            return s // Return as is
        }
    },
}
result, err = gshttp.Get(`http://xxxx.api`).SetStreamFac(fac).Request(200).Result()
```

##### Split by regular expression
```go
fac := &stream.Reges{
    Reges: `\x00{4}|[\x00-\x1F]`, // Split by ASCII
    CallFunc: func(s string, err error) {
        // Split received message
    },
    FormatFunc: func(s []byte) []byte {
        if gstool.SContains(cast.ToString(s), []string{`ignore`}) { // This content is not aggregated to the result
            return []byte{} // Return content can be customized
        } else {
            return s // Return as is
        }
    },
}
result, err = gshttp.Get(`http://xxxx.api`).SetStreamFac(fac).Request(200).Result()
```

##### Split by fixed byte length
```go
fac := &stream.BytsNum{
    Num: 255, // Return by fixed byte length
    CallFunc: func(s string, err error) {
        // Split received message
    },
    FormatFunc: func(s []byte) []byte {
        if gstool.SContains(cast.ToString(s), []string{`ignore`}) { // This content is not aggregated to the result
            return []byte{} // Return content can be customized
        } else {
            return s // Return as is
        }
    },
}
result, err = gshttp.Get(`http://xxxx.api`).SetStreamFac(fac).Request(200).Result()
```


## 3. Concurrent Execution Channels
Multiple tasks can be added to the execution queue, you can choose to execute all or execute any one
```go
// Execute all
task := gstask.NewTask()
callBack := gstask.CallbackFunc{
    Func: func() *gstask.Result {
        // TODO Specific business execution, need to return gstask.Result type
    },
    Timeout: 3 * time.Second, // timeout
    Id:      0, // execution business id, this id will be returned regardless of timeout or success
}
task.Add(callBack)
task.Add(callBack1) // continue adding
resultList := task.RunAll() // execute all
result := task.RunOne() // execute, return the first completed or timed out result
```

## 4. NSQ Consumer and Publish Messages
NSQ does not use goroutine concurrency, following a single client can only handle one message at a time
```go
nsqConfig := gsnsq.NsqConfig{
    LookUpHost: "127.0.0.1:4161", // consumption address
    PubMsgHost: "127.0.0.1:4150", // publish message address
}
nsq := gsnsq.NsqStruct{
    Topic:   "event",
    Channel: "event",
    Config:  nsqConfig,
}
// Create consumer 1 means creating 1 consumer
_ = nsq.CreateConsumer(1, func(s string, att uint16) bool {
    // TODO Business execution s is the consumer message content, att is retry count, initially 1
    return true // return true to continue, return false to requeue
})
// Stop (all messages consumed)
nsq.ConsumerShutDown()
// Create publisher Built-in highest 1000 concurrent send buffer
_ = nsq.CreateProducer()
```

## 5. Redis
### Basic Configuration
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
        UserName: "xx",// High version Redis has account name
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

### Redis Lock (Waiting or Non-Waiting)
```go
lock := gslock.NewRedisLock(redisCli, time.Second*30)
// One-time lock value judgment and return
// bool Whether the lock was obtained
// string If the lock was not obtained, the current value of the lock
// error Exception
b, s, err := lock.GetLock(`lock key value`, `lock value`)

// Continuous lock value judgment and return
// maxTry Number of attempts
// wait Interval between lock attempts when lock not obtained
// breakFunc Interrupt lock method. Called before each lock attempt. If it returns a non-empty string, interrupt and return to the second string return value; this can be used to check if the lock already exists, and if so, interrupt and return finally
// Return values:
// bool Whether the lock was obtained
// string String returned by interrupt, other situations return empty string
// error Exception
b, s, err := lock.GetWaitLock(`lock key value`, `lock value`, 10, time.Second, breakFunc)
```

### Cache Quick Build for Usage
- func() can be replaced with database query operations
- When encountering type mismatch, it will return the empty value of the input type
- Generic supports common types
#### Hash storage map
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

#### String storage map
- func() can be replaced with database query operations
- When encountering type mismatch, it will return the empty value of the input type
- Generic supports common types
```go
data, err := gsdb.RedisGetMapString(client, `test2`, func() (map[int]map[any]any, error) {
    return map[int]map[any]any{
        2: {
            "a": 1,
        },
    }, nil
}, time.Hour)

```

## 6. SSH
### One-time Command
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

### Interactive
#### Interactive execution of a command, receiving all returns, waiting for command to complete
```go
sshTerminal := gsssh.NewSshTerminal(gsssh.NewSsh(&gsssh.SshConfig{
    Host:     "11.11.11.11",
    Port:     "22",
    UserName: "xxx",
    Password: "xxxx",
}))
// Setting callback here will receive all returns including SSH connection welcome messages
sshTerminal.SetFuncStreamReceive(func(s string) {
    gstool.FmtPrintlnLogTime(`Received content %s`, s)
})
// This can be called multiple times without creating a new connection
// Execute a task to check current folder situation, maximum 5 seconds, if not completed then return first but SetFuncStreamReceive will continue receiving
ret, err := sshTerminal.RunCommandWait(`ls -l`, time.Second*5)
if err != nil {
    fmt.Println(err.Error())
    return
}
gstool.FmtPrintlnLogTime(`Final result %s`, ret)

```

#### Interactive execution of a command, receiving command return (excluding SSH connection system messages), waiting for command to complete
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
// Set callback at this time, then only subsequent command returns will be received
sshTerminal.SetFuncStreamReceive(func(s string) {
    gstool.FmtPrintlnLogTime(`Received content %s`, s)
})
// Execute a task to check current folder situation, maximum 5 seconds, if not completed then return first but SetFuncStreamReceive will continue receiving
// Can be used to continuously monitor command returns, for example: tail -f /var/log/test/log continuously returns new content
// If the current interactive session is only for one command, then call RunCommand, then it will not block returning ret, all results are received through SetFuncStreamReceive
ret, err := sshTerminal.RunCommandWait(`ls -l`, time.Second*5)
if err != nil {
    fmt.Println(err.Error())
    return
}
gstool.FmtPrintlnLogTime(`Final result %s`, ret)
```

## 8. Aliyun OSS
See oss_client and oss_quick in gsali for details