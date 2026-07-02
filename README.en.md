# go-tool

[中文](./README.md)

## Installation

```shell
go get -u github.com/w896736588/go-tool
```

## 1. mysql/pgsql/sqlite3

### 1. Create connection

```go
// The second parameter true means enabling automatic field type conversion
// SSH is optional
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
// Enable debug hook, all executed SQL will output full logs
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

### 2. Quick operations

#### Description

- Quick operations are designed to make simple SQL easier to write in a chained style, especially for cases like `like` and `in`. For complex SQL, use native SQL queries.
- When `autoTrans` is enabled, condition values and update values will be automatically converted to the corresponding types based on the table schema. MySQL itself also supports this.
- Supported operators: `> < >= <= <> like in not in between`

#### Examples

- Update all records that match the conditions

```go
upNum, queryErr := mysql.QuickUpdate(`tbl_user`, map[string]any{
    `id`: [20,21], // Automatically converted to in
	`username`: []any{`like` , `%1`},
    `username1`: []any{`not in` , []any{1,2,`3`}},
    `username1`: []any{`in` , []any{1,2,3}},
    `create_time`: []any{`between` , []any{12312,23422}},
    `rawsql#1`: []any{`rawsql`, `role_id1 = ?`, []any{`39`}}, // Define rawsql#1 as a unique string; precompiled values here will not be type-converted automatically
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

- Special handling for pgsql inserts to get the new ID

```go
id , queryErr := mysql.QuickInsert(`tbl_user`, map[string]any{} , `id`)
```

- Use transactions, with support for passing them into multiple database operations

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

- Quick query for joins and similar connections

```go
q := msql.QuickQuery(`tbl_test1 r`, `r.id`, map[string]any{}).
    Join(`left join tbl_test2 u on u.id = r.xx and u.xxx = ?`, `xxxxx`).
    Join(`....`)
ret, err := q.One()
gstool.FmtPrintlnLog(`%s`, q.GetSql())
```

- Extract results into a slice

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

- Extract results into a map

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

- Convert results into a map

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

### 3. Get the actual executed SQL logs

- Get the log for a single execution

```go
q := mysql.QuickSelect(`tbl_user`, []string{
	`id` : []any{`>` , 1}
}).OffsetLimit(0 , 100)
_,_ := q.All()
sql : q.GetSql() // Get the actual executed SQL
```

- Register a global log hook

```go
client.RegisterDebugHook(func(sql string, err error) {
    gstool.FmtPrintlnLogTime(`sql %s`, sql)
    gstool.FmtPrintlnLogTime(`error %v`, err)
})
```

### 4. Automatic type conversion notes

- Automatic conversion includes values for operations such as `in`, `not in`, and other operators, including query conditions and update values.
- Automatic conversion will query the table schema. If a field does not exist, the schema will be refreshed. If a field type changes, the service needs to be restarted; otherwise it may cause type conversion errors on insert and execution may fail.
- If a field's type changes, you need to restart the service, otherwise automatic conversion may use the wrong type.
- The second parameter of the `rawsql` type does not participate in automatic type conversion.

### 5. Other notes

- For fields of type `"date"`, `"datetime"`, and `"timestamp"`, pass a `string` instead of `time.Time`.

## 2. HTTP client

### 1. GET

```go
gshttp.Get(`http://xxxx/api`).Result()
```

### 2. POST

#### Submit arrays (`application/x-www-form-urlencoded` or `multipart/form-data`)

```go
// 1. First method: call BodyMap() multiple times to set multiple values for the same key; it will be converted to an array automatically
gshttp.PostForm(`http://xxxx.api`).
	BodyMap(map[string]any{}).BodyMap(map[string]any{}).Request(5 * time.Second).Result()
// 2. Second method: set an array parameter and it will be converted automatically
gshttp.PostForm(`http://xxxx.api`).
	BodyMap(map[string]any{
	`params` : []string{`a`, `b`}
}).Request(5 * time.Second).Result()
```

#### `application/json` request

```go
gshttp.PostJson(`http://xxxx.api`).
	BodyStr(`{"appid" : 1}`).Result()
```

#### `application/x-www-form-urlencoded` request

```go
gshttp.PostForm(`http://xxxx.api`).
	BodyMap(map[string]any{}).Request(5 * time.Second).Result()
```

#### Submit as `multipart/form-data`, with support for uploading multiple files

```go
gshttp.PostMultiForm(`http://xxxx.api`).BodyMap(map[string]any{}).
	BodyFile(`file` , `local path` , `xxx.png`).
	BodyFile(`file` , `local path` , `xxx.png`).Request(5 * time.Second).Result()
```

#### Example: upload a file to WeChat and get the material ID

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
        return nil,errors.New(fmt.Sprintf(`failed to get file extension`))
    }
    targetFileName := gstool.Md5(cast.ToString(fileContent)) + ext
    // Temporary storage directory
    localFilePath := `/tmp/` + targetFileName
    crErr := gstool.FilePutContent(localFilePath, cast.ToString(fileContent))
    if crErr != nil {
        return nil,crErr
    }
    ret, err := gshttp.PostMultiForm(url).BodyFile(`media`, localFilePath, fileName).Request(20 * time.Second).ResultStr()
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

Use `SetAllowHttpStatus` to specify a list of allowed HTTP status codes. If the response status code is not in the list, an error will be returned.
Default behavior: only 200 is allowed. When `allowHttpStatus` is set, response status codes within the list will not produce an error, and the response body will be parsed normally.

```go
// Allow 200 and 204 status codes; returning 204 will not cause an error and will return an empty body as usual
gshttp.Get(`http://xxxx/api`).SetAllowHttpStatus(200, 204).Result()
// Allow 404; even if the resource does not exist, no error will be thrown, and you can parse the error response body yourself
gshttp.Get(`http://xxxx/api`).SetAllowHttpStatus(200, 404).Result()
// Without SetAllowHttpStatus, only 200 is allowed by default
gshttp.Get(`http://xxxx/api`).Result()
```

#### keep-alive

When enabled, connections to the same target IP and port are reused, avoiding another TCP handshake. Disabled by default.

```go
gshttp.Get(`http://xxxx/api`).OpenKeepAlive().Result()
```

#### Receive as a stream

##### Split by string

```go
fac := &stream.Byts{
    Byts: []byte("\n\n"),
        CallFunc: func(s string, err error) {
        h.StreamMsg(s, false)
    },
    FormatFunc: func(s []byte) []byte {
        if gstool.SContains(cast.ToString(s), []string{`ignore`}) { // This content will not be merged into the result
            return []byte{} // You can define the returned content yourself
        } else {
            return s // Return as-is
        }
    },
}
result, err = gshttp.Get(`http://xxxx.api`).SetStreamFac(fac).Request(200 * time.Second).Result()
```

##### Split by regular expression

```go
fac := &stream.Reges{
    Reges: `\x00{4}|[\x00-\x1F]`, // Split by ascii
    CallFunc: func(s string, err error) {
        // Message obtained after splitting
    },
    FormatFunc: func(s []byte) []byte {
        if gstool.SContains(cast.ToString(s), []string{`ignore`}) { // This content will not be merged into the result
            return []byte{} // You can define the returned content yourself
        } else {
            return s // Return as-is
        }
    },
}
result, err = gshttp.Get(`http://xxxx.api`).SetStreamFac(fac).Request(200 * time.Second).Result()
```

##### Split by fixed byte length

```go
fac := &stream.BytsNum{
    Num: 255, // Return by fixed byte length
    CallFunc: func(s string, err error) {
        // Message obtained after splitting
    },
    FormatFunc: func(s []byte) []byte {
        if gstool.SContains(cast.ToString(s), []string{`ignore`}) { // This content will not be merged into the result
            return []byte{} // You can define the returned content yourself
        } else {
            return s // Return as-is
        }
    },
}
result, err = gshttp.Get(`http://xxxx.api`).SetStreamFac(fac).Request(200 * time.Second).Result()
```

## 3. Concurrent execution channel

Multiple tasks can be added to the execution queue, and you can choose to execute all of them or any one of them.

```go
// Execute all
task := gstask.NewTask()
callBack := gstask.CallbackFunc{
    Func: func() *gstask.Result {
        // TODO The actual business logic to execute must return a gstask.Result
    },
    Timeout: 3 * time.Second, // Timeout
    Id:      0, // ID of the task; returned whether it times out or succeeds
}
task.Add(callBack)
task.Add(callBack1) // Continue adding more
resultList := task.RunAll() // Execute all
result := task.RunOne() // Execute and return the first completed or timed-out result
```

## 4. NSQ consume and publish messages

NSQ does not use goroutine concurrency. One client can handle only one message at a time.

```go
nsqConfig := gsnsq.NsqConfig{
    LookUpHost: "127.0.0.1:4161", // Consumer address
    PubMsgHost: "127.0.0.1:4150", // Publisher address
}
nsq := gsnsq.NsqStruct{
    Topic:   "event",
    Channel: "event",
    Config:  nsqConfig,
}
// Create consumer; 1 means create one consumer
_ = nsq.CreateConsumer(1, func(s string, att uint16) bool {
    // TODO Business logic. s is the message content, att is the retry count, starting from 1
    return true // Return true to continue, return false to requeue
})
// Stop (after all messages are consumed)
nsq.ConsumerShutDown()
// Create publisher, with a built-in buffer for up to 1000 concurrent sends
_ = nsq.CreateProducer()
```

## 5. Redis

### Basic configuration

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
        UserName: "xx",// Newer Redis versions support account names
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

### Redis lock (waiting or non-waiting)

```go
lock := gslock.NewRedisLock(redisCli, time.Second*30)
// Check the lock once and return
// bool: whether the lock was acquired
// string: if not acquired, the current value of the lock
// error: exception
b, s, err := lock.GetLock(`lock key`, `lock value`)

// Keep trying to acquire the lock and return
// maxTry: number of attempts
// wait: interval between attempts when the lock is not acquired
// breakFunc: called before each lock attempt; if it returns a non-empty string, the locking process is interrupted and that string is returned as the second return value. This can be used to detect that the lock already exists and stop early.
// Return values:
// bool: whether the lock was acquired
// string: the string returned on interruption; empty string in other cases
// error: exception
b, s, err := lock.GetWaitLock(`lock key`, `lock value`, 10, time.Second, breakFunc)
```

### Build cache quickly for use

- `func()` can be replaced with a database query
- If the type does not match, an empty value of the input type will be returned
- Generics support common types

#### Store a map as hash

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

#### Store a map as string

- `func()` can be replaced with a database query
- If the type does not match, an empty value of the input type will be returned
- Generics support common types

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

### One-time command

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

#### Run one command interactively, receive all output, and wait for the command to finish

```go
sshTerminal := gsssh.NewSshTerminal(gsssh.NewSsh(&gsssh.SshConfig{
    Host:     "11.11.11.11",
    Port:     "22",
    UserName: "xxx",
    Password: "xxxx",
}))
// Setting the callback here will receive all output, including the SSH connection welcome message
sshTerminal.SetFuncStreamReceive(func(s string) {
    gstool.FmtPrintlnLogTime(`Received content %s`, s)
})
// This can be called multiple times without recreating the connection
// Run a command to inspect the current directory, wait up to 5 seconds; if it does not finish, return first, but SetFuncStreamReceive will continue receiving output
ret, err := sshTerminal.RunCommandWait(`ls -l`, time.Second*5)
if err != nil {
    fmt.Println(err.Error())
    return
}
gstool.FmtPrintlnLogTime(`Final result %s`, ret)
```

#### Run one command interactively, receive command output only (excluding SSH system messages), and wait for the command to finish

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
// If you set the callback now, it will only receive output from subsequent commands
sshTerminal.SetFuncStreamReceive(func(s string) {
    gstool.FmtPrintlnLogTime(`Received content %s`, s)
})
// Run a command to inspect the current directory, wait up to 5 seconds; if it does not finish, return first, but SetFuncStreamReceive will continue receiving output
// This can be used to continuously monitor command output, for example `tail -f /var/log/test/log`
// If the current interactive session is used for only one command, you can call RunCommand so it does not block waiting for ret, and all results are received through SetFuncStreamReceive
ret, err := sshTerminal.RunCommandWait(`ls -l`, time.Second*5)
if err != nil {
    fmt.Println(err.Error())
    return
}
gstool.FmtPrintlnLogTime(`Final result %s`, ret)
```

## 8. Aliyun OSS

See `oss_client` and `oss_quick` in `gsali` for details.
