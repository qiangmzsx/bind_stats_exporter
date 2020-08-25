package idgen

import (
	"fmt"
	"log"
	"math"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"gitlab.luojilab.com/zeroteam/common/bit"
	"gitlab.luojilab.com/zeroteam/common/ip"
)

// uint64 一个id 由3倍分组成， 最高位的timestamp ,中位的自增序列seq 及低位的ip
// timestamp+seq+ip

// 基于内网ip 与时间戳的一个id 生成器

// 同一个子网内产生的id 不重复
// 公司 docker 分配的ip 为172.19 的b 类子网，占用20 位
// 时间戳默认30 位(可用20 年)，则可用于自增的倍数为:64-20-30=14 位 , 即1s 内最多允许产生 2^14=16384 个id
// 公司 docker 分配的ip 为10.6 的a 类子网，占用24 位
// 时间戳默认30 位(可用20 年)，则可用于自增的倍数为:64-24-30=10 位 , 即1s 内最多允许产生 2^8=1024 个id
// 目前主要给trace 生成traceid spanid, 只要保证一段时间内（如一年内）id 不冲突即可
//
const (
	// 所存储的时间是一个时间差是 当前时间与baseTimeStamp 时间差对应的秒
	// 你用的时候可以调整baseTimeStamp到当前时间 ，但是正式使用之后就不要调整此值了，否则新生成的
	// id可能会与老id冲突
	BaseTimeStamp = "2020-04-20 02:45:59" // 格式里的0不能少， 比如03 不能写成3 否则解析会出错
	// 如果仅使用1年，则25位就够了， 可以根据需要 适当调整此数值
	TimestampBitCount = 30 // 存到秒级 30 位可保20 年内不重复
	// uint64除最高位不用，及timestampBits外 ，63-32=31 位， 还有空余的31位可用
	// 假如1秒内最大有可能生成1024个id 的话， 则用于自增序列sequence 所占的位数就是10
// 1月= 30*24*3600s= 2#100111 10001101 00000000 24位
// 24 天              2#11111 10100100 00000000 23位
// 1年=365*24*3600s=2#1 11100001 00110011 10000000 25 位
// 10年=             s    2#10010 11001100 00000011 00000000 29 位够用10 年
// 20年:  2#100101 10011000 00000110 00000000  #30 位
// 50年=  2#1011101 11111100 00001111 00000000 #31 位

)

type IDGenOption func(*IDGen)

// WithUpdateTimestamp 为true,此种算法生成的id 中会含有当前时间戳的信息, 故可保证不同机器产生的id 基本是有序递增的,
// (同一秒内不同机器产生的id 不是有序递增 )，缺点是，每秒可产生的id 数有限
// 若为false, 则不主动更新时间戳，只有当可用seq 用尽后，timestamp 位置才会递增，优点是可以容忍某1 秒内产生的id>seq 允许的最大值，
// 只要平均每秒的id 数量<seq 允许的最大值即可，缺点：不同机器产生的id, 不保证有序递增
func WithUpdateTimestamp(updateTimestamp bool) IDGenOption {
	return func(idgen *IDGen) {
		idgen.updateTimestamp = updateTimestamp
	}
}

// WithTimestampBitCount 表示timestamp 占多少位，默认30 位可产生20 年内不会重复的id
func WithTimestampBitCount(timestampBitCount int) IDGenOption {
	return func(idgen *IDGen) {
		if timestampBitCount > 32 || timestampBitCount <= 0 {
			panic("timestampBitCount should not >32 and <=0")
		}
		idgen.timestampBitCount = uint64(timestampBitCount)
	}
}

// WithBaseTimestamp 存入id 中的timpstamp 将等于time.Now().Unix()=baseTimestamp
// 默认为 BaseTimeStamp 常量所定义
func WithBaseTimestamp(baseTimestamp int64) IDGenOption {
	return func(idgen *IDGen) {
		idgen.baseTimestamp = uint64(baseTimestamp)
	}
}

// 手动指定ip值，及占用的位数
func WithIP(ip uint64, ipBitCount int) IDGenOption {
	return func(idgen *IDGen) {
		idgen.ip = uint64(ip)
		idgen.ipBitCount = uint64(ipBitCount)
	}
}

// NewIDGen id 生成器
// WithUpdateTimestamp 为true,此种算法生成的id 中会含有当前时间戳的信息, 故可保证不同机器产生的id 基本是有序递增的,
// (同一秒内不同机器产生的id 不是有序递增 )，缺点是，每秒可产生的id 数有限
// 若为false, 则不主动更新时间戳，只有当可用seq 用尽后，timestamp 位置才会递增，优点是可以容忍某1 秒内产生的id>seq 允许的最大值，
// 只要平均每秒的id 数量<seq 允许的最大值即可，缺点：不同机器产生的id, 不保证有序递增
// WithTimestampBitCount 表示timestamp 占多少位，默认30 位可产生20 年内不会重复的id
func NewIDGen(options ...IDGenOption) *IDGen {
	idgen := &IDGen{
		timestampBitCount: TimestampBitCount,
		baseTimestamp:     defaultBaseTimestamp,
	}
	idgen.initDefaultIP()
	for _, opt := range options {
		opt(idgen)
	}

	idgen.timestamp = idgen.getNow()
	idgen.seqBitCount = 64 - idgen.ipBitCount - idgen.timestampBitCount
	if idgen.seqBitCount <= 0 {
		panic(fmt.Sprintf("ip 段占了%d 位, 时间戳占了%d 位 无空余 bit位,用于自增序列",
			idgen.ipBitCount, idgen.timestampBitCount))
	}
	if Debug {
		log.Printf("idgen timestamp_bit_count=%d,timestamp=%d,ip_bit_count=%d,ip=%d,seq_bit_count=%d,seq=%d,max_id_each_second=%d",
			idgen.timestampBitCount, idgen.timestamp, idgen.ipBitCount, idgen.ip, idgen.seqBitCount, idgen.seq, int(math.Pow(2, float64(idgen.seqBitCount))))
	}

	return idgen
}

type IDGen struct {
	ip         uint64 // ip 占对应的值，如c 类ip 只需要存ip 对应的低16 位+ 2 位的用于区分 a b c 类ip，
	ipBitCount uint64

	timestampBitCount uint64
	timestamp         uint64

	seq         uint64
	seqBitCount uint64

	updateTimestamp bool
	baseTimestamp   uint64
	mutex           sync.Mutex
}

func (idgen *IDGen) initDefaultIP() {
	myIPString := getMyIP()
	myIP := net.ParseIP(myIPString)
	if ip.IpBetween(net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255"), myIP) {
		// C类地址：192.168.0.0--192.168.255.255
		// 2#1100000010101000 00000000 00000000
		// 2#1100000010101000 11111111 11111111
		// c 类地址只需要存低16 位即可
		ipLow := bit.BitInt(ip.Ipv42u32(myIPString)).GetUint16(0) // 取ip 的低16 位
		// var ipInfo bit.BitInt
		// ipInfo.SetUint16(0, ipLow)
		// ipInfo.SetByte(2, 0) // 	// 表示是c 类地址 占两位 ,
		// idgen.ip = ipInfo.ToUint64()
		idgen.ip = uint64(ipLow)
		idgen.ipBitCount = 16 // 则用于自增的有14 位 2^14  ,  一秒内最大产生16384 个id
	} else if ip.IpBetween(net.ParseIP("172.16.0.0"), net.ParseIP("172.31.255.255"), myIP) {
		// B类地址：172.16.0.0--172.31.255.255
		// 2#101011000001 0000 00000000 00000000
		// 2#101011000001 1111 11111111 11111111
		// 只需要存取低20 位
		ipLow := bit.BitInt(ip.Ipv42u32(myIPString)).GetValue(0, 20) // 取ip 的低20 位
		// var ipInfo bit.BitInt
		// ipInfo.SetValue(0, 20, ipLow)
		// // ipInfo.SetValue(20, 2, 1) // 表示是b 类地址, 占两位，
		// idgen.ip = ipInfo.ToUint64()
		idgen.ip = ipLow
		idgen.ipBitCount = 20 // 则用于自增的有10 位 2^10 65535 ,  一秒内最大产生1024 个id
	} else if ip.IpBetween(net.ParseIP("10.0.0.0"), net.ParseIP("10.255.255.255"), myIP) {
		// A类地址：10.0.0.0--10.255.255.255
		// 2#1010 00000000 00000000 00000000
		// 2#1010 11111111 11111111 11111111
		ipLow := bit.BitInt(ip.Ipv42u32(myIPString)).GetValue(0, 24) // 取ip 的低24 位
		// var ipInfo bit.BitInt
		// ipInfo.SetValue(0, 24, ipLow)
		// // ipInfo.SetValue(24, 2, 2) // 表示是a 类地址, 占两位
		// idgen.ip = ipInfo.ToUint64()
		idgen.ip = ipLow
		idgen.ipBitCount = 24 // 则用于自增的有6 位 2^6 64 ,  一秒内最大产生64 个id
	} else { // 否存存全ip
		idgen.ip = uint64(ip.Ipv42u32(myIPString))
		idgen.ipBitCount = 32
	}

}
func (idgen *IDGen) getTimestampAndSeq(now uint64) (timestamp, seq uint64) {
	if !idgen.updateTimestamp {
		// 这种情况只更新seq 使用atomic 就够了，不必加锁
		seq = atomic.AddUint64(&idgen.seq, 1)
		return idgen.timestamp, seq
	}

	idgen.mutex.Lock()
	defer idgen.mutex.Unlock()
	if now != idgen.timestamp {
		idgen.timestamp = now
		idgen.seq = 0
	} else {
		idgen.seq += 1
	}
	return idgen.timestamp, idgen.seq
}

func (idgen *IDGen) getTimestampSeq() (timestampSeq uint64) {
	now := idgen.getNow()
	timestamp, seq := idgen.getTimestampAndSeq(now)
	timestampSeq = (timestamp << idgen.seqBitCount) + seq
	timestamp = (timestampSeq >> idgen.seqBitCount)
	if timestamp > now {
		log.Printf("idgen id 生成过快，所用时间戳超卖，当前时间戳：%d, 所用时间戳：%d", now, timestamp)
	}

	return

}
func (idgen *IDGen) Generate() uint64 {
	seq := idgen.getTimestampSeq()
	return seq<<idgen.ipBitCount + idgen.ip
}
func (idgen *IDGen) GenerateHex() string {
	return strconv.FormatUint(idgen.Generate(), 16)
}

var mockNow uint64

func (idgen *IDGen) getNow() uint64 {
	if mockNow != 0 {
		return mockNow
	}
	return uint64(time.Now().Unix()) - idgen.baseTimestamp
}

var mockIP string

func getMyIP() string {
	if mockIP == "" {
		return ip.GetInnerIP()
	}
	return mockIP

}

var defaultBaseTimestamp uint64

func init() {
	idGenBaseTimestamp, err := time.Parse("2006-01-02 15:04:05", BaseTimeStamp)
	if err != nil {
		fmt.Println("err 解析baseTimeStamp发生错误 ，时间格式不对", BaseTimeStamp, "应该形如2006-01-02 15:04:05")
		panic("err 解析baseTimeStamp发生错误")
	}
	defaultBaseTimestamp = uint64(idGenBaseTimestamp.Unix())

}

var Debug bool
