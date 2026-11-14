package agent

import (
	"crypto/sha256"
	"fmt"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInstallerDigestVerification(t *testing.T) {
	script := renderInstallScript("token", "https://scheduler.example")
	start := strings.Index(script, "verify_package() {")
	if start < 0 {
		t.Fatal("missing verifier")
	}
	end := strings.Index(script[start:], "\nif [ \"$OS\" = \"windows\" ]; then")
	if start < 0 || end < 0 {
		t.Fatal("missing verifier")
	}
	verifier := script[start : start+end]
	path := filepath.Join(t.TempDir(), "package.tar.gz")
	content := []byte("test package")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		digest string
		ok     bool
	}{
		{fmt.Sprintf("%x", sha256.Sum256(content)), true},
		{strings.Repeat("0", 64), false},
		{"invalid", false},
		{"", true}, // Legacy installations warn rather than break.
	} {
		cmd := exec.Command("bash", "-c", "set -e\n"+verifier+"\nverify_package "+shellQuote(path))
		cmd.Env = append(os.Environ(), "GOCRON_NODE_SHA256="+tc.digest)
		out, err := cmd.CombinedOutput()
		if (err == nil) != tc.ok {
			t.Fatalf("digest result %v: %s", err, out)
		}
	}
}

func TestCanonicalServerURL(t *testing.T) {
	for _, value := range []string{"https://scheduler.example", "http://localhost:5920/", "https://[::1]:5920"} {
		if _, err := canonicalServerURL(value); err != nil {
			t.Fatal(err)
		}
	}
	for _, value := range []string{"javascript:alert(1)", "https://user:pass@example.com", "https://example.com/a", "https://example.com?x=1", "https://example.com#frag", "//example.com"} {
		if _, err := canonicalServerURL(value); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
	t.Setenv("GOCRON_PUBLIC_URL", "https://scheduler.example/")
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "http://attacker.example/api/agent/install.sh", nil)
	if got := getServerURL(c); got != "https://scheduler.example" {
		t.Fatalf("trusted host header: %s", got)
	}
}

func TestShellQuoteRoundTrip(t *testing.T) {
	for _, value := range []string{"http://localhost:5920", "https://[::1]:5920", "http://evil$(printf INJECTED)", "http://evil'; printf INJECTED; #", "a\nb`printf INJECTED`", ""} {
		t.Run(value, func(t *testing.T) {
			out, err := exec.Command("sh", "-c", "printf '%s' "+shellQuote(value)).Output()
			if err != nil {
				t.Fatal(err)
			}
			if string(out) != value {
				t.Fatalf("shell changed literal: got %q, want %q", out, value)
			}
		})
	}
}

func TestInstallScriptTreatsInputsAsLiterals(t *testing.T) {
	for _, serverURL := range []string{"http://localhost:5920", "http://evil$(printf INJECTED)", "http://evil'; printf INJECTED; #"} {
		token := "token'$(printf INJECTED)"
		script := renderInstallScript(token, serverURL)
		cmd := exec.Command("bash", "-n")
		cmd.Stdin = strings.NewReader(script)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("invalid script: %v: %s", err, out)
		}
		// Run only the generated assignments, never the installer itself.
		var assignments []string
		for _, line := range strings.Split(script, "\n") {
			if strings.HasPrefix(line, "TOKEN=") || strings.HasPrefix(line, "GOCRON_SERVER=") {
				assignments = append(assignments, line)
			}
		}
		if len(assignments) != 2 {
			t.Fatalf("got %d assignments", len(assignments))
		}
		out, err := exec.Command("bash", "-c", strings.Join(assignments, "\n")+"\nprintf '%s\\n%s' \"$TOKEN\" \"$GOCRON_SERVER\"").Output()
		if err != nil || string(out) != token+"\n"+serverURL {
			t.Fatalf("installer changed literals: %q, %v", out, err)
		}
	}
}
