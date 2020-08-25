package api

import (
	"context"
	"errors"
	"time"
)

type Valuable interface {
	IsNil() bool                          // 判断 存不存在
	Exists(optionalSubKey ...string) bool // 判断子key 存不存在 val.Exists("sub.sub2.sub3")

	// GetFileType() string // ".toml", ".json", ".yaml", ".yml"
	WithFileType(filetype string) Valuable
	Raw() (str string)

	// ToString return string value.
	// demo: `{"name":{"first":"Janet","last":"Prichard"},"age":47}`
	// WithFileType("json").ToString("name.first") return `Janet`
	// ToString() return `{"name":{"first":"Janet","last":"Prichard"},"age":47}`
	ToString(optionalSubKey ...string) (str string, err error)
	MustString(optionalSubKey ...string) (str string)
	DefaultString(defaultValue string, optionalSubKey ...string) (str string)

	ToBool(optionalSubKey ...string) (bool, error)
	MustBool(optionalSubKey ...string) bool
	DefaultBool(defaultBool bool, optionalSubKey ...string) bool

	ToFloat64(optionalSubKey ...string) (float64, error)
	MustFloat64(optionalSubKey ...string) float64
	DefaultFloat64(defaultValue float64, optionalSubKey ...string) float64

	ToUint64(optionalSubKey ...string) (uint64, error)
	MustUint64(optionalSubKey ...string) uint64
	DefaultUint64(defaultValue uint64, optionalSubKey ...string) uint64

	ToInt64(optionalSubKey ...string) (int64, error)
	MustInt64(optionalSubKey ...string) int64
	DefaultInt64(defaultValue int64, optionalSubKey ...string) int64

	ToInt(optionalSubKey ...string) (int, error)
	MustInt(optionalSubKey ...string) int
	DefaultInt(defaultValue int, optionalSubKey ...string) int

	ToInt32(optionalSubKey ...string) (int32, error)
	MustInt32(optionalSubKey ...string) int32
	DefaultInt32(defaultValue int32, optionalSubKey ...string) int32

	ToDuration(optionalSubKey ...string) (time.Duration, error)
	MustDuration(optionalSubKey ...string) time.Duration
	DefaultDuration(defaultValue time.Duration, optionalSubKey ...string) time.Duration

	Unmarshal(result interface{}, optionalSubKey ...string) error


	GetUpdateTime() int64 // 内部用
}

type Backend interface {
	Get(ctx context.Context, key string, opts ...QueryOption) (value Valuable, err error)
	GetAndWatch(ctx context.Context, key string, onChange func(newValue Valuable), opts ...QueryOption) (value Valuable, err error)
	BackendName() string
	Alias(aliasKey, realkey string) Backend
}

type QueryOptions struct {
	FileType     string
	RemotePrefix string // prefix for consul
	FilePrefix   string // prefix for filebackend
	EnvPrefix    string // prefix for envbackend
	Token        string // 对加密key 中
}
type QueryOption func(*QueryOptions)

// WithFileType 支持的filetype见 utils.SupportedFileTypes
func WithFileType(fileType string) QueryOption {
	return func(opts *QueryOptions) {
		if fileType != "" {
			opts.FileType = fileType
		}

	}
}
func WithRemotePrefix(consulPrefix string) QueryOption {
	return func(opts *QueryOptions) {
		opts.RemotePrefix = consulPrefix
	}
}
func WithEnvPrefix(envPrefix string) QueryOption {
	return func(opts *QueryOptions) {
		opts.EnvPrefix = envPrefix
	}
}
func WithFilePrefix(consulPrefix string) QueryOption {
	return func(opts *QueryOptions) {
		opts.FilePrefix = consulPrefix
	}
}

// WithToken 对于一些加密数据，必须提供passwd 才能获取相应的key
func WithToken(token string) QueryOption {
	return func(opts *QueryOptions) {
		opts.Token = token
	}
}

var ErrorNotFound = errors.New("value not found")
