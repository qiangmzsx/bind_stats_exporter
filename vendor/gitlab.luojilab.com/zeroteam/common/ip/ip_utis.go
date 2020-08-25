package ip

import (
	"bytes"
	"errors"
	"net"
	"strings"
)

func Ipv42i32(ipv4 string) int32 {
	ipint := net.ParseIP(ipv4).To4()
	var i32ip int32 = 0
	i32ip |= int32(ipint[0]) << 24
	i32ip |= int32(ipint[1]) << 16
	i32ip |= int32(ipint[2]) << 8
	i32ip |= int32(ipint[3])
	return i32ip
}

func Ipv42u32(ipv4 string) uint32 {
	ipint := net.ParseIP(ipv4).To4()
	var i32ip uint32 = 0
	i32ip |= uint32(ipint[0]) << 24
	i32ip |= uint32(ipint[1]) << 16
	i32ip |= uint32(ipint[2]) << 8
	i32ip |= uint32(ipint[3])
	return i32ip
}

// moved from artemis/util to here
func GetInternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "127.0.0.1", errors.New("failed to get IP")
}

func GetInnerIP() (ip string) { // 获取第一个内网ip
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip := GetIpFromAddress(addr.String())
			if ip == "" {
				continue
			}
			if IsInnerIP(ip) {
				return ip
			}
		}
	}
	return

}

// 10.0.0.0-10.255.255.255
// 172.16.0.0—172.31.255.255
// 192.168.0.0-192.168.255.255
func IsInnerIP(ipStr string) bool {
	// if ipStr == "127.0.0.1" {
	// 	return true
	// }
	// if ipStr == "::1" { // ipv6
	// 	return true
	// }
	if strings.HasPrefix(ipStr, "FEC0:") { // ipv6 site local,就像IPV4中的私网保留地址一样
		return true
	}
	if strings.HasPrefix(ipStr, "FE80:") { // ipv6 link local
		return true
	}

	ip := net.ParseIP(ipStr)
	if IpBetween(net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255"), ip) {
		return true
	}
	if IpBetween(net.ParseIP("172.16.0.0"), net.ParseIP("172.31.255.255"), ip) {
		return true
	}
	if IpBetween(net.ParseIP("10.0.0.0"), net.ParseIP("10.255.255.255"), ip) {
		return true
	}
	return false
}

func IsLOIP(ipStr string) bool { // 127.0.0.1
	if ipStr == "127.0.0.1" {
		return true
	}
	if ipStr == "::1" { // ipv6
		return true
	}

	return false
}

// IpBetween(net.ParseIP(from), net.ParseIP(to), net.ParseIP(test))
//test to determine if a given ip is between two others (inclusive)
func IpBetween(from net.IP, to net.IP, test net.IP) bool {
	if from == nil || to == nil || test == nil {
		// glog.Warning("An ip input is nil") // or return an error!?
		return false

	}
	from16 := from.To16()
	to16 := to.To16()
	test16 := test.To16()
	if from16 == nil || to16 == nil || test16 == nil {
		// glog.Warning("An ip did not convert to a 16 byte") // or return an error!?
		return false

	}

	if bytes.Compare(test16, from16) >= 0 && bytes.Compare(test16, to16) <= 0 {
		return true

	}
	return false

}

func GetIpFromAddress(address string) string { // address="ip:port"
	list := strings.Split(address, ":")
	if len(list) == 2 {
		return list[0]
	}
	list = strings.Split(address, "/")
	if len(list) == 2 {
		return list[0]
	}
	return ""
}

// GetLocalIP 获得本机IP ,强烈推荐使用这个函数，而不是GetInnerIP()
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
