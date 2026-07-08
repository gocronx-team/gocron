package llm

import "testing"

func TestValidateBaseURL(t *testing.T) {
	valid := []string{
		"https://api.openai.com/v1",
		"http://127.0.0.1:1234/v1",      // local model (LM Studio) must be allowed
		"http://ollama.internal:11434",  // internal host must be allowed
		"  https://api.example.com/v1 ", // surrounding whitespace tolerated
	}
	for _, u := range valid {
		if err := ValidateBaseURL(u); err != nil {
			t.Errorf("ValidateBaseURL(%q) = %v, want nil", u, err)
		}
	}

	invalid := []string{
		"",                   // empty
		"api.openai.com/v1",  // no scheme
		"ftp://example.com",  // wrong scheme
		"file:///etc/passwd", // dangerous scheme, no host
		"gopher://x",         // wrong scheme
		"https://",           // no host
		"://nohost",          // unparliamentary
	}
	for _, u := range invalid {
		if err := ValidateBaseURL(u); err == nil {
			t.Errorf("ValidateBaseURL(%q) = nil, want error", u)
		}
	}
}
