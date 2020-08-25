package env

import (
	"fmt"
	"strconv"

	"gitlab.luojilab.com/zeroteam/common/ip"
)

var envInstance *Env = &Env{envMode: ModeDev}

func GetEnv() *Env {
	return envInstance
}

type Env struct {
	dcid             string // env DC_ID
	hostname         string // env HOSTNAME
	appName          string
	debug            bool
	envMode          EnvMode
	localIP          string // 本机内网ip
	httpPort         uint16 // artemis启动后 填充此值
	componentId      string // env COMPONENT_ID,如 bj-1.prod.bauhinia
	home             string
	serviceUid       string // env SERVICE_UID zeroteam/ddarticle/default
	downwardNodeIP   string
	downwardNodeName string
	downwardCPULimit int // docker分配的cpu核数 DOWNWARD_CPU_LIMIT
}

func (this *Env) SetDownwardCPULimit(value int) *Env {
	this.downwardCPULimit = value
	return this
}
func (this Env) GetDownwardCPULimit() int {
	return this.downwardCPULimit
}

func (this *Env) SetDownwardNodeIP(value string) *Env {
	this.downwardNodeIP = value
	return this
}
func (this Env) GetDownwardNodeIP() string {
	return this.downwardNodeIP
}
func (this *Env) SetDownwardNodeName(value string) *Env {
	this.downwardNodeName = value
	return this
}
func (this Env) GetDownwardNodeName() string {
	return this.downwardNodeName
}

func (this *Env) SetServiceUid(value string) *Env {
	this.serviceUid = value
	return this
}
func (this Env) GetServiceUid() string {
	return this.serviceUid
}

func (this Env) GetHome() string {
	return this.home
}

func (this *Env) SetDcid(value string) *Env {
	this.dcid = value
	return this
}
func (this Env) GetDcid() string {
	return this.dcid
}
func (this *Env) SetHostname(value string) *Env {
	this.hostname = value
	return this
}
func (this Env) GetHostname() string {
	return this.hostname
}
func (this *Env) SetAppName(value string) *Env {
	this.appName = value
	return this
}
func (this Env) GetAppName() string {
	return this.appName
}
func (this *Env) SetDebug(value bool) *Env {
	this.debug = value
	return this
}
func (this Env) GetDebug() bool {
	return this.debug
}
func (this Env) GetEnvMode() EnvMode {
	return this.envMode
}
func (this *Env) SetLocalIP(value string) *Env {
	this.localIP = value
	return this
}
func (this *Env) SetHttpPort(value uint16) *Env {
	this.httpPort = value
	return this
}
func (this Env) GetHttpPort() uint16 {
	return this.httpPort
}
func (this Env) GetHttpPortAsString() string {
	return strconv.Itoa(int(this.httpPort))
}
func (this *Env) SetComponentId(value string) *Env {
	this.componentId = value
	return this
}
func (this Env) GetComponentId() string {
	return this.componentId
}

func (e Env) String() string {
	return fmt.Sprintf("DCID:%s,hostname:%s,appName:%s,debug:%v,env:%s,httpPort:%d,componentId=%s",
		e.dcid, e.hostname, e.appName, e.debug, e.envMode.String(), e.httpPort, e.componentId)
}
func (this *Env) GetLocalIP() string {
	if this.localIP != "" {
		return this.localIP
	}
	localIP, err := ip.GetInternalIP()
	if err != nil && localIP != "" {
		this.SetLocalIP(localIP)
	}

	return this.localIP
}
func (e *Env) SetEnvMode(str string) error {
	return e.envMode.Set(str)
}

type EnvMode string

// ENV_RUN
const (
	ModeDefault    EnvMode = ""
	ModeDev        EnvMode = "development"
	ModeTesting    EnvMode = "testing"
	ModeSimulation EnvMode = "simulation"
	ModeProduction EnvMode = "production"
)

func (m EnvMode) IsProduction() bool {
	return m == ModeProduction
}

func (m EnvMode) IsSimulation() bool {
	return m == ModeSimulation
}
func (m EnvMode) IsDefault() bool {
	return m == ModeDefault
}
func (m EnvMode) IsDev() bool {
	return m == ModeDev
}

func (m EnvMode) IsTesting() bool {
	return m == ModeTesting
}
func (m EnvMode) String() string {
	return string(m)
}
func (m EnvMode) Chinese() string {
	if m.IsDev() {
		return "dev环境"
	} else if m.IsTesting() {
		return "测试环境"
	} else if m.IsProduction() {
		return "线上环境"
	} else if m.IsSimulation() {
		return "仿真环境"
	} else if m.IsDefault() {
		return "默认环境"
	}
	return m.String()
}

// Set sets the value of the named command-line flag.
func (m *EnvMode) Set(str string) error {
	if str == "" {
		return nil
	}

	*m = EnvMode(str)
	if str == "dev" {
		*m = ModeDev
	}
	if str == "online" {
		*m = ModeProduction
	}
	if str == "test" {
		*m = ModeTesting
	}
	if str == "simu" {
		*m = ModeSimulation
	}
	if *m != ModeDev && *m != ModeSimulation && *m != ModeProduction && *m != ModeTesting {
		return fmt.Errorf("unsupport envMode %s", str)
	}
	return nil
}
