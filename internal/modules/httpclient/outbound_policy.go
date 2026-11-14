package httpclient

import (
	"context"
	"errors"
	"net"
	"os"
	"time"
)

// Opt-in at process startup. Legacy internal HTTP tasks remain supported by default.
var publicOnly = os.Getenv("GOCRON_HTTP_PUBLIC_ONLY") == "1"

func publicAddress(ip net.IP) bool {
	return ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback() &&
		!ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() &&
		!blockedSpecialAddress(ip)
}

// Translation/tunneling ranges can encode a non-public IPv4 destination even
// though Go classifies their IPv6 representation as global unicast.
var specialNetworks = func() []*net.IPNet {
	var networks []*net.IPNet
	for _, cidr := range []string{
		"100.64.0.0/10", "0.0.0.0/8", "192.0.0.0/24", "198.18.0.0/15",
		"192.0.2.0/24", "198.51.100.0/24", "203.0.113.0/24", "240.0.0.0/4",
		"::/96", "64:ff9b::/96", "64:ff9b:1::/48", "2002::/16", "2001::/32", "2001:db8::/32",
	} {
		_, network, _ := net.ParseCIDR(cidr)
		networks = append(networks, network)
	}
	return networks
}()

func blockedSpecialAddress(ip net.IP) bool {
	for _, network := range specialNetworks {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// Resolve and validate at connection time, then dial the numeric IP so a second
// DNS lookup cannot switch to a private address. Redirects use this same dialer.
func outboundDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	if !publicOnly {
		return dialer.DialContext(ctx, network, address)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 || len(ips) > 64 {
		return nil, errors.New("outbound DNS address count is invalid")
	}
	for _, ip := range ips {
		if ip.Zone != "" || !publicAddress(ip.IP) {
			return nil, errors.New("outbound private or special-use address blocked by public-only policy")
		}
	}
	return dialValidatedAddresses(ctx, network, port, ips, dialer.DialContext)
}

// Connection fallback only; no HTTP request has been sent at this point.
func dialValidatedAddresses(ctx context.Context, network, port string, ips []net.IPAddr, dial func(context.Context, string, string) (net.Conn, error)) (net.Conn, error) {
	var lastErr error
	for i, ip := range ips {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		attempt := ctx
		cancel := func() {}
		if deadline, ok := ctx.Deadline(); ok {
			attempt, cancel = context.WithTimeout(ctx, time.Until(deadline)/time.Duration(len(ips)-i))
		}
		conn, err := dial(attempt, network, net.JoinHostPort(ip.IP.String(), port))
		cancel()
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
