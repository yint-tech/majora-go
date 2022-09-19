package infra

import (
	"net"

	"iinti.cn/majora-go/log"
)

const (
	netname = "ppp0"
)

func GetPPP() string {
	return GetIpByName(netname)
}

func GetIpByName(netname string) string {
	ni, err := net.InterfaceByName(netname)

	if err != nil {
		log.Run().Warnf("get %s ip error %s", netname, err)
		return ""
	}

	addrs, err := ni.Addrs()

	if err != nil {
		log.Run().Warnf("get ip addr err %s", err)
		return ""
	}

	if len(addrs) == 0 {
		log.Run().Warnf("get ip addr empty ")
		return ""
	}

	var ipv4Addr net.IP

	for _, addr := range addrs {
		if ipv4Addr = addr.(*net.IPNet).IP.To4(); ipv4Addr != nil {
			break
		}
	}

	if ipv4Addr == nil {
		log.Run().Warnf("interface %s don't have an ipv4 address", netname)
		return ""
	}
	return ipv4Addr.String()
}
