package util

import (
	"net"
	"net/http"
	"strings"
)

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

// GetClientRealIP 获取客户端的真实IP地址
func GetClientRealIP(r *http.Request) (string, bool) {
	headers := []string{
		"X-Forwarded-For",
		"Proxy-Client-IP",
		"WL-Proxy-Client-IP",
		"X-Real-Ip",
	}

	for _, header := range headers {
		ip := r.Header.Get(header)
		if ip != "" {
			// X-Forwarded-For 可能包含多个 IP 地址 用逗号分隔 取第一个有效的
			ips := strings.Split(ip, ",")
			for _, ipPart := range ips {
				ipPart = strings.TrimSpace(ipPart)
				if isValidIp(ipPart) {
					return ipPart, true
				}
			}
		}
	}

	// 如果没有获取到有效的 IP，则返回远程地址
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip, isValidIp(ip)
}

// isValidIp 校验 IP 地址是否有效
func isValidIp(ip string) bool {
	return net.ParseIP(ip) != nil
}
