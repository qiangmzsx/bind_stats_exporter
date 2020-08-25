package model

type EndPoint struct {
	Ipv4        int32  `json:"ipv4"`
	ServiceName string `json:"servicename"`
	Port        int16  `json:"port"`
	IPv6        string `json:"ipv6""` //not support yet
	DCID        string `json:"dcid"`
}
