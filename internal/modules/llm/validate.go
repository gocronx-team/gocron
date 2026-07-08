package llm

import (
	"errors"
	"net/url"
	"strings"
)

// ErrInvalidBaseURL 表示 base_url 不是合法的 http/https 地址。
var ErrInvalidBaseURL = errors.New("llm: base_url must be an absolute http(s) URL with a host")

// ValidateBaseURL 校验大模型接口地址:必须是带 host 的绝对 http/https URL。
// 这道边界校验把"未受控的用户输入"收敛为受控值,拒绝 file://、gopher:// 等
// 危险 scheme(SSRF 面收敛)。注意:出于支持本地/自建模型(如 127.0.0.1、内网
// Ollama)的需要,不限制 host 指向内网——该能力是管理员显式配置的、且管理员本就
// 有全系统控制权,不构成额外提权。
func ValidateBaseURL(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ErrInvalidBaseURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrInvalidBaseURL
	}
	if u.Host == "" {
		return ErrInvalidBaseURL
	}
	return nil
}
