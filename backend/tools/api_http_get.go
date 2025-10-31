package tools

import (
	"net"
	"net/http"
	"strings"
)

// Fetch Session from Request Context
// Expects UseSession to be present in the http handler chain otherwise it panics
func GetSession(r *http.Request) *SessionData {
	v := r.Context().Value(SESSION_KEY)
	if v == nil {
		panic("missing session in request context; this request should have returned earlier")
	}
	return v.(*SessionData)
}

// Get IP Address of Incoming Client
func GetRemoteIP(r *http.Request) string {
	remoteAddr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}
	clientIP := net.ParseIP(remoteAddr)

	// Skip if request was not proxied
	if len(HTTP_IP_HEADERS) == 0 || len(HTTP_IP_PROXIES) == 0 {
		return clientIP.String()
	}

	// Walk through headers in configured order (most recent first)
	// Scan Headers as Configured by Environment
	for _, header := range HTTP_IP_HEADERS {
		hv := r.Header.Get(header)
		if hv == "" {
			continue
		}
		for _, ipStr := range strings.Split(hv, ",") {
			ipStr = strings.TrimSpace(ipStr)
			ip := net.ParseIP(ipStr)
			if ip != nil && !isTrustedProxy(ip) {
				return ip.String()
			}
		}
	}

	// Proxy is misconfigured, use fallback!
	return clientIP.String()
}

// proxy helper :p

func isTrustedProxy(ip net.IP) bool {
	for _, cidr := range HTTP_IP_PROXIES {
		if _, network, err := net.ParseCIDR(cidr); err == nil {
			if network.Contains(ip) {
				return true
			}
		} else if proxyIP := net.ParseIP(cidr); proxyIP != nil && proxyIP.Equal(ip) {
			return true
		}
	}
	return false
}
