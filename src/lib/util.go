package lib

import (
	"net"
	"os"
)

type IPVersion int

const (
	IPv4 IPVersion = iota
	IPv6
)

func GetLocalIP(ver IPVersion) string {
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		switch ver {
		case IPv4:
			if ip4 := addr.To4(); ip4 != nil {
				return ip4.String()
			}
		case IPv6:
			if ip6 := addr.To16(); ip6 != nil {
				return ip6.String()
			}
		}
	}
	return ""
}
