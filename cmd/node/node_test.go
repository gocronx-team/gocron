package main

import "testing"

func TestSecurityModePreservesLegacyAndRequiresTLS(t *testing.T) {
	for _, tc := range []struct {
		strict, tls, wantError bool
	}{{false, false, false}, {false, true, false}, {true, false, true}, {true, true, false}} {
		if err := validateSecurity(tc.strict, tc.tls); (err != nil) != tc.wantError {
			t.Fatalf("strict=%v tls=%v: %v", tc.strict, tc.tls, err)
		}
	}
}
