# 主要功能：

1. 默认日志打印到标准输出
2. verbose日志实现，参考glog实现，可通过设置不同verbose级别，默认不开启
3. 支持json 格式日志
4. 支持结构化日志
5. 支持从context中获取traceId 等信息并输出到日志，方便链路追踪

# 用法
## 基本用法
```go
import (
	"gitlab.luojilab.com/zeroteam/common/log"
)
log.Info("hello this is a log.")
log.Warningf("hello %s",name)
log.Errorln("hello ","world")
```

## 支持结构化日志 ,需要使用者确保参数成对存在，且key为string
```go
//输出key1=value1 key2=value2 或 {"key1":"value1","key2":"value2"},根据输出格式而定
log.InfoKV("key1","value1","key2","value2")
log.WarningKV("key1","value1","key2","value2")
log.ErrorKV("key1","value1","key2","value2")
```
## 固定输出特定的KV
以下是log库内定的一些特殊Key

| key      | 对应变量    | 描述                                                 |
|----------+-------------+------------------------------------------------------|
| LEVEL   | KeyLevel    | 日志等级                                             |
| TIME    | KeyTime     | 日志时间戳                                           |
| FILE    | KeySource   | 日志对应文件位置                                     |
| MSG     | KeyMsg      | 日志内容,log.Info("hello")对应{"MSG":"hello"}       |
| APP     | KeyAppname  | 日志app名                                            |
| HOST    | KeyHostName | hostname                                             |
| TRACEID | KeyTraceId  | traceId                                              |
| ENV     | KeyEnvMode  | 部署环境如development,production, simulation,testing |
| DCID    | KeyDCID     | 机房id                                               |
默认日志固定输出以下指定的key
```go
defaultFixedKeys             = []string{KeyLevel, KeyTime, KeyTraceId, KeySource}
```
可通过以下方式指定固定输出哪些key,如固定输出X-Uid,X-D等key.
```go
log.SetLogFixedKeys(log.KeyLevel, log.KeyTime, log.KeyTraceId, log.KeySource,"X-Uid","X-D")
```
对于log库内定的这些key ,其对应的值有些从环境变量中取，有些从启动参数里取，如KeyHostName,KeyAppname等，
对于KeyTraceid及其他用户自定义的Key(如"X-Uid"),则需要从ctx中取,具体用法如下，用户需要使用带ctx的日志打印方式

```go
// 注意此处可以是系统的 "context"包,也可以是"gitlab.luojilab.com/zeroteam/common/context"
ctx:=context.Background()
ctx=context.WithValue(ctx,traceable.HeaderTraceId,"this_is_traceid")//traceid比较特殊，其对应的key为traceable.HeaderTraceId即"X-Trace-Id"
ctx=context.WithValue(ctx,"X-Uid","this_is_uid")

log.ContextInfof(ctx,"hello %s",name)//输出 {"_traceid":"this_is_traceid","X-Uid":this_is_uid,"_msg":"hello name"}及其他key
log.ContextWarning(ctx,"hello %s",name)
log.ContextErrorKV(ctx,"key1","value1")
```
## 更方便的使用方式是使用	"gitlab.luojilab.com/zeroteam/common/context"
context包直接提供了日志打印接口，且常用的X-Uid,traceid等key在artemis中已进行填充，可直接使用
```go
// 获取到context.IContext后用法如下
ctx.Info("hello this is a log with context support,I can get extra info from context")
ctx.Errorf("hello this is a log with context support,I can get extra info from context")
ctx.WarningKV("key1","value1","key2","value2")

// 另外提一点zeroteam/artemis/engine.Context直接继承自gitlab.luojilab.com/zeroteam/common/context.IContext
// 故 使用最新artemis的项目，升级artemis与common库后，可以以下使用方式使用log库
import (
	"gitlab.luojilab.com/zeroteam/artemis/engine"
	"gitlab.luojilab.com/zeroteam/common/context"
)

 func (this *HelloEndpoint) SayHello(c engine.Context) {//c中已经在middleware里填充的X-Trace-Id,X-Uid等信息
  c.Info("this is a log")
  c.WarningKV("key1","value1","key2","value2")
   GetUserService().SaveUser(c,user)
 }
 func (this *UserService) SaveUser(ctx common.IContext,user User) {
  ctx.Info("this is a log",user)
 }
 //所以建议但凡出现 common/context.IContext/或 artemis/engine.Context的地方都建议使用这种方式
 //也建议你的函数的首参数都以 context.IContext开始
 ```
 另外说一点 gitlab.luojilab.com/zeroteam/common/context.IContext继承了系统自带的"context".Context
 故，所有接收"context".Context的地方都可以传common/context.IContext

# 支持调整Depth
```
log.InfoDepth(1,"hello this is a log.")
log.WarningfDepth(1,"hello this is a log.")
log.ContextWarningfDepth(ctx,1,"hello this is a log.")
```
# verbose日志实现
verbose日志的理解，可以参考ssh命令加了 -v -vv -vvv的不同

v越多的时候，输出的debug信息越详细

比如下面当设置V的等级为2的时候 log.V(3).Info()的输出是不打印出来的
 ```go
 log.SetV(2) #设置全局v的等级

if log.V(1) { log.Info("log this") }
if log.V(2) { log.Info("log this") }
if log.V(3) { log.Info("log this") }
log.V(2).Info("log this")
log.V(2).ContextInfof(ctx,"hello %s","world")

```
# 可单独配置每个文件的verbose级别
可以针对特定的文件 开启verbose级别日志
```
m:=map[string]int{
"user.go":2,
"service*":1,//所有文件名中在service的文件，日志级别为1
}
log.SetModule(m)
```

# 日志配置

1. 启动参数 or 环境变量

| 	启动参数		           | 环境变量		                | 说明                                                                                                              |
| ----------                     | ---                             | ---                                                                                                               |
| 	log.logtostderr	        | LOG_TO_STDERR	               | 是否开启标准输出,default:true                                                                                     |
| 	log.json                   | LOG_FORMAT_JSON                 | 	是否以json格式打印日志：true/false,default:true                                                               |
| 	log.dir		            | LOG_DIR		                 | 文件日志路径,为空则不打印到文件,default 空                                                                        |
| 	log.file_prefix	        | LOG_FILE_PREFIX		         | 文件日志名前缀,当指定此参数时，此时会按文件大小滚动分隔，并以此为文件名前缀                                       |
| 	log.file_name	          | LOG_FILE_NAME		           | 文件日志名（不包含路径的纯文件名）,当指定此参数时，日志文件名将固定不变，不再按文件大小滚动分隔                                |
| 	log.write_to_same_file	 | LOG_WRITE_TO_SAME_FILE 		 | 是否不区分日志等级，都写到同一个文件，如果log.file_name非空，则此值无意义                                         |
| 	log.symlinks	           | LOG_SMYLINKS	                | 将日志文件软链到此处，格式如:dir1/,dir2/,dir3/filename.log ,以逗号分隔，目录名以/结尾，否则将被当作文件名而非目录 |
| 	log.filter	             | LOG_FILTER	                  | 配置需要过滤的字段：field1,field2                                                                                 |
| 	log.fixed_keys	         | LOG_FIXED_KEYS	              | 配置指定的key固定出现在日志中，其value从ctx中获取                                                                 |
| 	log.v		              | LOG_V		                   | verbose日志级别,回忆下ssh root@host -v -vv -vvv 的不同就可理解                                                    |
| 	log.module	             | LOG_MODULE	                  | 可单独配置每个文件的verbose级别：file=1,file2=2,service=1,dao*.go=2                                               |
用法
```
flag.Parse()
log.Init()
// 用法 调用flag.Parse()之后，再调用log.Init()
// 此时 才能解析到命令行的flag参数
// Init之后 再使用log.Info()等，配置才能生效
```


2. 或者通过以下方式定制logger

```go
logConf := log.NewConfig().
SetAsDefaultLogger(true). //将此logger作为默认logger使用
SetSymlinks("/tmp/", "/var/log/","/tmp/s/filename.log").//设置把最新的日志文件软链接到何处,目录需以/结尾，否则认为其包含了文件名
SetLogFormatJson(true)//日志格式:json ,只对文件日志有效，控制台不支持json格式
.SetFixedKeys(log.KeyDCID,"X-Uid")	// 始终打印dcid,及X-uid到日志中,其对应的值从ctx中取
if dir != "" {
  logConf.SetDir(dir) //输出到文件
}

if env.GetEnv().GetDebug() {
    logConf.SetLogToStderr(true) //控制是否输出到stderr
} else {
    logConf.SetLogToStderr(false)
}
logConf.SetFilePrefix(env.GetEnv().GetAppName())//日志文件的前缀
log.NewLogger(logConf) // 根据logConf的配置创建一个新logger

```
# 压测数据
```go
//性能:  glog>zeroteam/common/log >zaplog
//allocs: glog>zaplog>zeroteam/common/log

// BenchmarkTestGLOGInfo-4                  3000000              1244 ns/op             184 B/op          2 allocs/op
// BenchmarkTestZaplog-4                     500000             11657 ns/op             536 B/op         10 allocs/op
// BenchmarkTestZaplogJson-4                 300000             10659 ns/op             464 B/op          7 allocs/op

// BenchmarkTestInfo-4                      1000000              3040 ns/op            1160 B/op         12 allocs/op
// BenchmarkTestContextInfo-4               1000000              3339 ns/op            1208 B/op         15 allocs/op
// BenchmarkTestInfoJson-4                  1000000              4095 ns/op            2425 B/op         20 allocs/op
// BenchmarkTestContextInfoJson-4           1000000              4439 ns/op            2506 B/op         24 allocs/op
```
# 写日志文件相关：DropIfBufferFull参数
用于控制日志buffer满时，是否丢弃日志,以免buffer继续扩容占用更多内存
# 写日志文件相关：FlushInterval 参数
定时刷盘时间间隔
# 写日志文件相关：FileBufferSize参数
用来控制写文件时的缓冲区大小，glog给的大小是256k,

var FileBufferSize = 1024 * 1024
比如下面4C8G的压测机序列写 约466M/s

# 硬盘序列写性能
```
yum install libaio libaio-devel fio
```

测试序列写， 每次写大小256k的块，默认的ioengine=sync,运行60s(runtime),-direct=1(跳过buffer,直接刷盘)
```
fio -filename=/tmp/hello -direct=1 -bs=256k -rw=write -size=5G -numjobs=8 -runtime=60 -group_reporting -name=test_for_sync
#结果 ：write: IOPS=413, BW=103MiB/s (108MB/s)(6204MiB/60023msec)
```
测试序列写， 每次写大小256k的块，默认的ioengine=sync,运行60s(runtime),-direct=0使用buffer
```
fio -filename=/tmp/hello -direct=0 -bs=256k -rw=write -size=5G -numjobs=8 -runtime=60 -group_reporting -name=test_for_sync
＃结果:write: IOPS=1777, BW=444MiB/s (466MB/s)(26.0GiB/60024msec)
```
