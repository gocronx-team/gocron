package utils

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// ClientIP returns c.ClientIP() with IPv6 loopback and IPv4-mapped-IPv6 forms
// collapsed to their plain IPv4 equivalents so audit / login logs read naturally.
//
// Examples:
//
//	::1                -> 127.0.0.1
//	::ffff:10.0.0.5    -> 10.0.0.5
//	192.168.1.10       -> 192.168.1.10 (unchanged)
func ClientIP(c *gin.Context) string {
	return NormalizeIP(c.ClientIP())
}

// NormalizeIP collapses IPv6 loopback and IPv4-mapped IPv6 addresses to their
// IPv4 equivalents. Returns the original string unchanged when it is empty
// or cannot be parsed as an IP.
func NormalizeIP(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}

	// Strip bracketed form like "[::1]:12345" that can leak through some proxies.
	if strings.HasPrefix(raw, "[") {
		if host, _, err := net.SplitHostPort(raw); err == nil {
			raw = host
		}
	}

	ip := net.ParseIP(raw)
	if ip == nil {
		return raw
	}

	// Collapse IPv6 loopback to IPv4 loopback.
	if ip.IsLoopback() && ip.To4() == nil {
		return "127.0.0.1"
	}

	// Prefer the IPv4 representation when available (covers ::ffff:a.b.c.d).
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}

	return ip.String()
}
