package util

import "net"

func GetAllIP() []*net.IPNet {
	ips := make([]*net.IPNet, 0)
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			ips = append(ips, ip)
		}
	}
	return ips
}

func GetIPv4() []string {
	ips := make([]string, 0)
	for _, ip := range GetAllIP() {
		if ip.IP.To4() != nil {
			ips = append(ips, ip.IP.String())
		}
	}
	return ips
}

func GetIPv6() []string {
	ips := make([]string, 0)
	for _, ip := range GetAllIP() {
		if ip.IP.To16() != nil {
			ips = append(ips, ip.IP.String())
		}
	}
	return ips
}
