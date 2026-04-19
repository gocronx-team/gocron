package utils

import "testing"

func TestNormalizeIP(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"::1", "127.0.0.1"},
		{"127.0.0.1", "127.0.0.1"},
		{"::ffff:127.0.0.1", "127.0.0.1"},
		{"::ffff:10.0.0.5", "10.0.0.5"},
		{"192.168.1.10", "192.168.1.10"},
		{"2001:db8::1", "2001:db8::1"},
		{"[::1]:54321", "127.0.0.1"},
		{"", ""},
		{"not-an-ip", "not-an-ip"},
	}

	for _, tc := range tests {
		if got := NormalizeIP(tc.in); got != tc.want {
			t.Errorf("NormalizeIP(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
