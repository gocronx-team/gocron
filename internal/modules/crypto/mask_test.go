package crypto

import "testing"

func TestMaskSecrets(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		secrets []string
		want    string
	}{
		{"empty text", "", []string{"abcde"}, ""},
		{"no secrets", "hello world", nil, "hello world"},
		{"single", "token=abc123 done", []string{"abc123"}, "token=" + MaskPlaceholder + " done"},
		{"multiple occurrences", "abcde and abcde", []string{"abcde"}, MaskPlaceholder + " and " + MaskPlaceholder},
		{"empty secret ignored", "keep this", []string{""}, "keep this"},
		{
			"short values below min length are not masked",
			"1 ok true abcd",
			[]string{"1", "ok", "true", "abcd"},
			"1 ok true abcd",
		},
		{
			"overlapping prefers longer",
			"pass=supersecret123",
			[]string{"super", "supersecret123"},
			"pass=" + MaskPlaceholder,
		},
		{
			"multiple distinct secrets",
			"a=11111 b=22222",
			[]string{"11111", "22222"},
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
