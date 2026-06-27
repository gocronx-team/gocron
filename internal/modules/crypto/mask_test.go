package crypto

import "testing"

func TestMaskSecrets(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		secrets []string
		want    string
	}{
		{"empty text", "", []string{"a"}, ""},
		{"no secrets", "hello world", nil, "hello world"},
		{"single", "token=abc123 done", []string{"abc123"}, "token=" + MaskPlaceholder + " done"},
		{"multiple occurrences", "abc and abc", []string{"abc"}, MaskPlaceholder + " and " + MaskPlaceholder},
		{"empty secret ignored", "keep this", []string{""}, "keep this"},
		{
			"overlapping prefers longer",
			"pass=supersecret123",
			[]string{"super", "supersecret123"},
			"pass=" + MaskPlaceholder,
		},
		{
			"multiple distinct secrets",
			"a=1111 b=2222",
			[]string{"1111", "2222"},
			"a=" + MaskPlaceholder + " b=" + MaskPlaceholder,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskSecrets(tt.text, tt.secrets); got != tt.want {
				t.Errorf("MaskSecrets() = %q, want %q", got, tt.want)
			}
		})
	}
}
