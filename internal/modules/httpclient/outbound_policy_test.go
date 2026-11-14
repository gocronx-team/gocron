package httpclient

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestValidatedAddressFallback(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var attempts []string
	client, peer := net.Pipe()
	defer peer.Close()
	conn, err := dialValidatedAddresses(ctx, "tcp", "443", []net.IPAddr{{IP: net.ParseIP("2606:4700:4700::1111")}, {IP: net.ParseIP("8.8.8.8")}}, func(attempt context.Context, _, addr string) (net.Conn, error) {
		attempts = append(attempts, addr)
		if _, ok := attempt.Deadline(); !ok {
			t.Fatal("unbounded connection attempt")
		}
		if len(attempts) == 1 {
			return nil, errors.New("unreachable")
		}
		return client, nil
	})
	if err != nil || conn != client || len(attempts) != 2 {
		t.Fatalf("fallback failed: %v %v", attempts, err)
	}
	_ = conn.Close()
}

func TestPublicAddress(t *testing.T) {
	if publicAddress(net.ParseIP("::127.0.0.1")) {
		t.Fatal("accepted deprecated IPv4-compatible loopback")
	}
	for _, address := range []string{"127.0.0.1", "10.0.0.1", "172.16.0.1", "192.168.1.1", "169.254.169.254", "100.100.100.200", "::1", "fe80::1", "fc00::1", "::ffff:127.0.0.1", "0.0.0.0", "224.0.0.1", "64:ff9b::7f00:1", "64:ff9b:1::a00:1", "2002:7f00:1::", "2001:db8::1", "198.18.0.1", "192.0.0.1", "240.0.0.1"} {
		if publicAddress(net.ParseIP(address)) {
			t.Fatalf("accepted non-public IP %s", address)
		}
	}
	for _, address := range []string{"8.8.8.8", "2606:4700:4700::1111"} {
		if !publicAddress(net.ParseIP(address)) {
			t.Fatalf("rejected public IP %s", address)
		}
	}
}

func BenchmarkPublicAddress(b *testing.B) {
	ip := net.ParseIP("8.8.8.8")
	b.ReportAllocs()
	for b.Loop() {
		publicAddress(ip)
	}
}

func TestLegacyInternalHTTPStillWorks(t *testing.T) {
	old := publicOnly
	publicOnly = false
	t.Cleanup(func() { publicOnly = old })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("internal service"))
	}))
	defer server.Close()
	response := Get(server.URL, 1)
	if response.StatusCode != 200 || response.Body != "internal service" {
		t.Fatalf("legacy internal HTTP broken: %+v", response)
	}
}

func TestPublicOnlyDialRejectsPrivateAddress(t *testing.T) {
	old := publicOnly
	publicOnly = true
	t.Cleanup(func() { publicOnly = old })
	for _, address := range []string{"127.0.0.1:80", "[::1]:80", "169.254.169.254:80"} {
		conn, err := outboundDialContext(context.Background(), "tcp", address)
		if conn != nil {
			_ = conn.Close()
		}
		if err == nil {
			t.Fatalf("accepted private target %s", address)
		}
	}
}

// Exercise the shared HTTP transport, not just the address classifier. A new
// transport avoids reusing a legacy connection when switching test policy.
func TestPublicOnlyHTTPBlocksLocalTargets(t *testing.T) {
	old := publicOnly
	publicOnly = true
	t.Cleanup(func() { publicOnly = old })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("blocked target received an HTTP request")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	transport := &http.Transport{DialContext: outboundDialContext}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: time.Second}
	for _, target := range []string{server.URL, "http://localhost:" + strconv.Itoa(server.Listener.Addr().(*net.TCPAddr).Port)} {
		response, err := client.Get(target)
		if response != nil {
			_ = response.Body.Close()
		}
		if err == nil {
			t.Fatalf("public-only transport accepted %s", target)
		}
	}
}

// Compare the old transport's dialer with the compatible policy-off path using
// real loopback HTTP. No external DNS or network service is needed.
func BenchmarkOutboundHTTPCompatibility(b *testing.B) {
	old := publicOnly
	publicOnly = false
	b.Cleanup(func() { publicOnly = old })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	defer server.Close()
	for _, keepAlive := range []bool{true, false} {
		name := "reused"
		if !keepAlive {
			name = "new-connection"
		}
		baseline := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
		for _, variant := range []struct {
			name string
			dial func(context.Context, string, string) (net.Conn, error)
		}{{"baseline", baseline.DialContext}, {"policy-off", outboundDialContext}} {
			b.Run(name+"/"+variant.name, func(b *testing.B) {
				transport := &http.Transport{DialContext: variant.dial, DisableKeepAlives: !keepAlive}
				defer transport.CloseIdleConnections()
				client := &http.Client{Transport: transport, Timeout: time.Second}
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					response, err := client.Get(server.URL)
					if err != nil {
						b.Fatal(err)
					}
					_, err = io.Copy(io.Discard, response.Body)
					_ = response.Body.Close()
					if err != nil || response.StatusCode != http.StatusOK {
						b.Fatalf("HTTP status=%d err=%v", response.StatusCode, err)
					}
				}
			})
		}
	}
}
