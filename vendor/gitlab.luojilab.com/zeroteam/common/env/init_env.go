package env

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"log" // gitlab.luojilab.com/zeroteam/common/log依赖env,故此处不能使用common/log

	"runtime"

	"github.com/pelletier/go-toml"
	"gitlab.luojilab.com/zeroteam/common/detclib/api"
	"gitlab.luojilab.com/zeroteam/common/idgen"
	"gitlab.luojilab.com/zeroteam/common/ip"
)

var IsTestMode bool

func init() {
	var err error
	if envInstance.hostname, err = os.Hostname(); err != nil || envInstance.hostname == "" {
		envInstance.hostname = os.Getenv("HOSTNAME")
	}

	envInstance.envMode = ModeDefault
	addFlag(flag.CommandLine)
}

// Init
// Deprecated: by InitWithValuable
func Init(conf *toml.Tree) {
	debug := conf.GetDefault("local.debug", false).(bool)
	GetEnv().SetDebug(debug)
	if GetEnv().GetEnvMode() == ModeDefault { // 以环境变量优先RUN_ENV，如果已经有值不再读配置文件中的local.env
		// ModeDev        EnvMode = "development"
		// ModeTesting    EnvMode = "testing"
		// ModeSimulation EnvMode = "simulation"
		// ModeProduction EnvMode = "production"
		envTree := conf.GetDefault("local.env", "") //
		if envTree != nil {
			err := GetEnv().SetEnvMode(envTree.(string))
			if err != nil {
				log.Print("env.Init().SetEnvMode", err)
			}

		}
	}

	name, ok := conf.GetDefault("local.name", "").(string)
	if ok && name != "" && GetEnv().GetAppName() == filepath.Base(os.Args[0]) {
		GetEnv().SetAppName(name)
	}
	if GetEnv().componentId != "" {
		tokens := strings.Split(GetEnv().componentId, ".")
		if len(tokens) > 0 {
			GetEnv().SetAppName(tokens[len(tokens)-1])
		}

	}

	port, ok := conf.GetDefault("local.port", 0).(int64)
	if !ok || port == 0 {
		laddr := conf.GetDefault("local.address", "0").(string) // ":8080"
		if len(laddr) > 1 {
			idx := strings.LastIndex(laddr, ":")
			if idx >= 0 {
				laddr = laddr[idx+1:]
			}
			address, err := strconv.Atoi(laddr)

			if err != nil {
				port = 0
			}
			port = int64(address)
		}
	}
	GetEnv().SetHttpPort(uint16(port))

	if !IsTestMode {
		log.Print("env:", GetEnv().String())
		if debug {
			idgen.Debug = true
		}

	}

}
func InitWithValuable(conf api.Valuable) {
	if conf == nil {
		return
	}

	debug := conf.DefaultBool(false, "local.debug")
	GetEnv().SetDebug(debug)
	if GetEnv().GetEnvMode() == ModeDefault { // 以环境变量优先RUN_ENV，如果已经有值不再读配置文件中的local.env
		// ModeDev        EnvMode = "development"
		// ModeTesting    EnvMode = "testing"
		// ModeSimulation EnvMode = "simulation"
		// ModeProduction EnvMode = "production"
		envTree := conf.MustString("local.env")
		if envTree != "" {
			err := GetEnv().SetEnvMode(envTree)
			if err != nil {
				log.Print("env.Init().SetEnvMode", err)
			}

		}
	}

	name := conf.MustString("local.name")
	if name != "" && GetEnv().GetAppName() == filepath.Base(os.Args[0]) {
		GetEnv().SetAppName(name)
	}
	if GetEnv().componentId != "" {
		tokens := strings.Split(GetEnv().componentId, ".")
		if len(tokens) > 0 {
			GetEnv().SetAppName(tokens[len(tokens)-1])
		}

	}

	port := conf.MustInt64("local.port")
	if port == 0 {
		laddr := conf.MustString("local.address")
		if len(laddr) > 1 {
			idx := strings.LastIndex(laddr, ":")
			if idx >= 0 {
				laddr = laddr[idx+1:]
			}
			address, err := strconv.Atoi(laddr)

			if err != nil {
				port = 0
			}
			port = int64(address)
		}
	}
	GetEnv().SetHttpPort(uint16(port))

	if !IsTestMode {
		log.Print("env:", GetEnv().String())
		if debug {
			idgen.Debug = true
		}
	}

}
func addFlag(fs *flag.FlagSet) {
	// env
	bindString(fs, &envInstance.home, "", "HOME", "", "$HOME")
	bindString(fs, &envInstance.dcid, "dc_id", "DC_ID", "", "avaliable DC. or use DC_ID env variable, value: bj-0.dev etc.")
	bindValue(fs, &envInstance.envMode, "run_env", "RUN_ENV", "deploy env, value: development/testing/simulation/production")
	bindString(fs, &envInstance.appName, "local.name", "LOCAL.NAME", filepath.Base(os.Args[0]), "name: appname")
	bindString(fs, &envInstance.componentId, "component_id", "COMPONENT_ID", "",
		"COMPONENT_ID: docker环境有环境变量COMPONENT_ID,如 bj-1.prod.bauhinia")
	bindString(fs, &envInstance.serviceUid, "service_uid", "SERVICE_UID", "",
		"SERVICE_UID: docker环境有环境变量,如 zeroteam/ddarticle/default")
	bindString(fs, &envInstance.downwardNodeIP, "downward_node_ip", "DOWNWARD_NODE_IP", "",
		"DOWNWARD_NODE_IP: docker环境有此环境变量")
	bindString(fs, &envInstance.downwardNodeName, "downward_node_name", "DOWNWARD_NODE_NAME", "",
		"DOWNWARD_NODE_NAME: docker环境有此环境变量:cn-beijing.i-2ze4sityha87pb0tdryl")
	bindBool(fs, &envInstance.debug, "local.debug", "LOCAL.DEBUG", false, "debug or not : true or false")
	bindInt(fs, &envInstance.downwardCPULimit, "downward_cpu_limit", "DOWNWARD_CPU_LIMIT", runtime.NumCPU(), "cpu cors c ount")
	localIP, err := ip.GetInternalIP()
	if err == nil && localIP != "" {
		envInstance.SetLocalIP(localIP)
	}
	// var apmEnable string
	// bindString(fs, &apmEnable, "apm_enable", "APM_ENABLE", "", "enable apm or not")
	// if strings.ToLower(apmEnable) == "true" || apmEnable == "1" {
	// 	GetEnv().SetApmEnable(true)
	// }

	flag.Int64("local.port", 0, "local.port")
	flag.String("local.address", "", "local.address")
}

// 优先级 flag>env>default
func bindString(fs *flag.FlagSet, ptr *string, flagName, envName, defaultValue string, desc string) {
	if len(envName) > 0 {
		envValue := os.Getenv(envName)
		if envValue != "" {
			defaultValue = envValue
		}
		*ptr = defaultValue
	}
	if flagName != "" {
		fs.StringVar(ptr, flagName, defaultValue, desc)
	}
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
		*ptr = defaultValue
	}
	if flagName != "" {
		fs.BoolVar(ptr, flagName, defaultValue, desc)
	}
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
		*ptr = defaultValue
	}
	if flagName != "" {
		fs.IntVar(ptr, flagName, defaultValue, desc)
	}
}

// 优先级 flag>env>default
func bindValue(fs *flag.FlagSet, ptr flag.Value, flagName, envName string, desc string) {
	if len(envName) > 0 {
		envValue := os.Getenv(envName)
		if envValue != "" {
			ptr.Set(envValue)
		}
	}
	if flagName != "" {
		fs.Var(ptr, flagName, desc)
	}
}
