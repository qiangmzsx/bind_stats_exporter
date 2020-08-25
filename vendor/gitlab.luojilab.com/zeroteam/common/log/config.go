package log

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func NewConfig() *config {
	c := config{}
	c = defaultConfigAfterParseFlag.Clone()
	// c.logToStderr = false
	// c.asDefaultLogger = false
	return &c

}

// 全局只维护一个v
var globalV int // V Enable V-leveled logging at the specified level.
// The syntax of the argument is a map of pattern=N,
// where pattern is a literal file name (minus the ".go" suffix) or
// "glob" pattern and N is a V level. For instance:
// [module]
//   "service" = 1
//   "dao*" = 2
// sets the V level to 2 in all Go files whose names begin "dao".
// 用来设置不同模块V的等级
var globalModule verboseModule = verboseModule{} //// type verboseModule map[string]int32

// Config log config.
type config struct {
	asDefaultLogger  bool
	logToStderr      bool // 是否打印到stdout
	logFormatJson    bool // 是否以json的形式打印日志
	fixedKeys        stringSlice
	dir              string
	filePrefix       string      // 日志文件的前缀，此时支持日志文件滚动
	filter           stringSlice // Filter tell log handler which field are sensitive message, use * instead.
	symlinks         stringSlice // 日志文件软链接到哪些目录，注意这里的值如果以"/"结果，会被当成目录，否则会被当成absolute文件名,
	fileName         string      //固定的文件名， 不支持日志文件滚动
	writeToSameFile  bool        //是否根据日志等级写到不同的日志文件中,当fileName非空时，此值即恒为true才可
	logStats         bool        // 是否允计在日志中打印统计信息 以便查找日志库的性能
	dropIfBufferFull bool        // buffer满时，是否允许丢弃一部分日志
}

func (this *config) SetDropIfBufferFull(value bool) *config {
	this.dropIfBufferFull = value
	return this
}

func (this *config) SetLogStats(value bool) *config {
	this.logStats = value
	return this
}
func (this *config) SetWriteToSameFile(value bool) *config {
	if this.fileName != "" {
		this.writeToSameFile = true
	} else {
		this.writeToSameFile = value
	}

	return this
}
func (this config) IsWriteToSameFile() bool {
	if this.fileName != "" {
		return true
	}
	return this.writeToSameFile
}

func (this *config) SetFileName(value string) *config {
	this.fileName = value
	return this
}
func (this config) GetFileName() string {
	return this.fileName
}

func (this *config) SetSymlinkDirs(value ...string) *config {
	for idx, dir := range value {
		if !strings.HasSuffix(dir, "/") {
			dir += "/"
			value[idx] = dir
		}
	}
	this.symlinks = value
	return this
}
func (this *config) SetSymlinks(value ...string) *config {
	this.symlinks = value
	return this
}
func (this config) GetSymlinks() []string {
	return this.symlinks
}

func (this config) String() string {
	return fmt.Sprintf("asDefaultLogger:%v\nlogToStderr:%v\nlogFormatJson:%v\nfixedKeys:%v\ndir:%s\nfilePrefix:%s\nfilter:%v\nsymlinks:%v\nv=%v\nmodule=%v",
		this.asDefaultLogger, this.logToStderr, this.logFormatJson, this.fixedKeys, this.dir, this.filePrefix, this.filter, this.symlinks, globalV, globalModule)
}

func (this *config) SetAsDefaultLogger(value bool) *config {
	this.asDefaultLogger = value
	return this
}
func (this config) GetAsDefaultLogger() bool {
	return this.asDefaultLogger
}
func (this *config) SetLogToStderr(value bool) *config {
	this.logToStderr = value
	return this
}
func (this config) GetLogToStderr() bool {
	return this.logToStderr
}
func (this *config) SetLogFormatJson(value bool) *config {
	this.logFormatJson = value
	return this
}
func (this config) GetLogFormatJson() bool {
	return this.logFormatJson
}
func (this *config) SetFixedKeys(value ...string) *config {
	this.fixedKeys = value
	return this
}
func (this config) GetFixedKeys() []string {
	return this.fixedKeys
}
func (this *config) SetDir(value string) *config {
	this.dir = value
	return this
}
func (this config) GetDir() string {
	return this.dir
}
func (this *config) SetFilePrefix(value string) *config {
	this.filePrefix = value
	return this
}
func (this config) GetFilePrefix() string {
	return this.filePrefix
}
func (this *config) SetFilter(value []string) *config {
	this.filter = value
	return this
}
func (this config) GetFilter() []string {
	return this.filter
}

func (c config) Clone() config {
	c2 := c
	c2.fixedKeys = make([]string, len(c.fixedKeys))
	copy(c2.fixedKeys, c.fixedKeys)

	c2.filter = make([]string, len(c.filter))
	copy(c2.filter, c.filter)

	return c2
}

// addFlag init log from dsn.
func addFlag(fs *flag.FlagSet, conf *config) {
	bindInt(fs, &globalV, "log.v", "LOG_V", int(globalV), "log verbose level, or use LOG_V env variable.")

	bindBool(fs, &conf.logToStderr, "log.logtostderr", "LOG_TO_STDERR", conf.logToStderr,
		"log enable logToStderr or not, or use LOG_TO_STDERR env variable.")

	bindString(fs, &conf.dir, "log.dir", "LOG_DIR", "", "log file path, or use LOG_DIR env variable.")

	bindString(fs, &conf.filePrefix, "log.file_prefix", "LOG_FILE_PREFIX", "",
		"log file name prefix, or use LOG_FILE_PREFIX env variable. support log file rotation.")
	bindString(fs, &conf.fileName, "log.file_name", "LOG_FILE_NAME", "",
		`use this as logfilename ,and don't support log file rotation, or use LOG_FILE_NAME env variable.\n
do not use log.file_name and log.file_prefix at the same time`)
	bindBool(fs, &conf.writeToSameFile, "log.write_to_same_file", "LOG_WRITE_TO_SAME_FILE", conf.writeToSameFile,
		"write to same file for different log level.")

	bindBool(fs, &conf.logFormatJson, "log.json", "LOG_FORMAT_JSON", conf.logFormatJson,
		"log format json or not , or use LOG_FORMAT_JSON env variable  value:true/false.")

	bindValue(fs, &conf.symlinks, "log.symlinks", "LOG_SMYLINKS",
		`symlink log file to these directories, or use LOG_SMYLINKS env variable,
format: dir1/,dir2/,dir3/filename.log,如果是目录请以/结尾，否则会被当成日志文件的绝对路径`)

	bindValue(fs, &conf.filter, "log.filter", "LOG_FILTER",
		"log field for sensitive message, or use LOG_FILTER env variable, format: field1,field2.")

	bindValue(fs, &conf.fixedKeys, "log.fixed_keys", "LOG_FIXED_KEYS",
		"fixed keys , or use LOG_FIXED_KEYS env variable, format: key1,key2.")

	// format: -log.module file=1,file2=2
	fs.Var(&globalModule, "log.module",
		"log verbose for specified module, or use LOG_MODULE env variable, format: file=1,file2=2,service=1,*_dao.go=2")

}

// 优先级 flag>env>default
func bindString(fs *flag.FlagSet, ptr *string, flagName, envName, defaultValue string, desc string) {
	if len(envName) > 0 {
		envValue := os.Getenv(envName)
		if envValue != "" {
			defaultValue = envValue
		}
	}
	fs.StringVar(ptr, flagName, defaultValue, desc)
}

// 优先级 flag>env>default
func bindBool(fs *flag.FlagSet, ptr *bool, flagName, envName string, defaultValue bool, desc string) {
	if len(envName) > 0 {
		envValue := os.Getenv(envName)
		if envValue != "" {
			envValue, err := strconv.ParseBool(envValue)
			if err == nil {
				defaultValue = envValue
			}
		}
	}
	fs.BoolVar(ptr, flagName, defaultValue, desc)
}

// 优先级 flag>env>default
func bindInt(fs *flag.FlagSet, ptr *int, flagName, envName string, defaultValue int, desc string) {
	if len(envName) > 0 {
		envValue := os.Getenv(envName)
		if envValue != "" {
			envValue, err := strconv.ParseInt(envValue, 10, 64)
			if err == nil {
				defaultValue = int(envValue)
			}
		}
	}
	fs.IntVar(ptr, flagName, defaultValue, desc)
}

// 优先级 flag>env>default
func bindValue(fs *flag.FlagSet, ptr flag.Value, flagName, envName string, desc string) {
	if len(envName) > 0 {
		envValue := os.Getenv(envName)
		if envValue != "" {
			ptr.Set(envValue)
		}
	}
	fs.Var(ptr, flagName, desc)
}

type verboseModule map[string]int

type stringSlice []string

func (f *stringSlice) String() string {
	return fmt.Sprint(*f)
}

// Set sets the value of the named command-line flag.
// format: -log.filter key1,key2
func (f *stringSlice) Set(value string) error {
	for _, i := range strings.Split(value, ",") {
		*f = append(*f, strings.TrimSpace(i))
	}
	return nil
}

func (m verboseModule) String() string {
	// FIXME strings.Builder
	var buf bytes.Buffer
	for k, v := range m {
		buf.WriteString(k)
		buf.WriteString(strconv.FormatInt(int64(v), 10))
		buf.WriteString(",")
	}
	return buf.String()
}

// Set sets the value of the named command-line flag.
// format: -log.module file=1,file2=2
func (m verboseModule) Set(value string) error {
	for _, i := range strings.Split(value, ",") {
		kv := strings.Split(i, "=")
		if len(kv) == 2 {
			if v, err := strconv.ParseInt(kv[1], 10, 64); err == nil {
				m[strings.TrimSpace(kv[0])] = int(v)
			}
		}
	}
	return nil
}
