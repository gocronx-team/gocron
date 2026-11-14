package httpclient

// http-client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ResponseWrapper struct {
	StatusCode int
	Body       string
	Header     http.Header
}

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// 优化：使用全局 HTTP 客户端，复用连接池
var defaultClient = &http.Client{
	Timeout: 300 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DialContext:         outboundDialContext,
	},
}

var clientFactory = func(timeout int) httpDoer {
	// 使用默认超时（300秒）或未设置超时时，直接返回全局客户端
	if timeout <= 0 || timeout == 300 {
		return defaultClient
	}
	// 其他超时值：创建新客户端但复用 Transport（连接池）
	return &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: defaultClient.Transport,
	}
}

func Get(url string, timeout int) ResponseWrapper {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return createRequestError(err)
	}

	return request(req, timeout)
}

func PostParams(url string, params string, timeout int) ResponseWrapper {
	buf := bytes.NewBufferString(params)
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return createRequestError(err)
	}
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	return request(req, timeout)
}

func PostJson(url string, body string, timeout int) ResponseWrapper {
	buf := bytes.NewBufferString(body)
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return createRequestError(err)
	}
	req.Header.Set("Content-type", "application/json")

	return request(req, timeout)
}

// blockedHeaders 禁止用户设置的危险 Header
var blockedHeaders = map[string]bool{
	"host":                true,
	"transfer-encoding":   true,
	"content-length":      true,
	"connection":          true,
	"upgrade":             true,
	"proxy-authorization": true,
	"proxy-connection":    true,
	"te":                  true,
	"trailer":             true,
}

// IsBlockedHeader 检查 header 是否在黑名单中
func IsBlockedHeader(name string) bool {
	return blockedHeaders[strings.ToLower(strings.TrimSpace(name))]
}

// ValidateHeaders 校验 headers JSON 格式并检查黑名单，返回错误信息
func ValidateHeaders(headersJSON string) error {
	if strings.TrimSpace(headersJSON) == "" {
		return nil
	}
	var headers map[string]string
	if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
		return fmt.Errorf("invalid JSON format")
	}
	for k := range headers {
		if IsBlockedHeader(k) {
			return fmt.Errorf("header %q is not allowed", k)
		}
	}
	return nil
}

// SetCustomHeaders 为请求设置自定义 Header（JSON 格式: {"Key": "Value", ...}）
// 黑名单中的 Header 会被跳过并记录日志
func SetCustomHeaders(req *http.Request, headersJSON string) {
	if strings.TrimSpace(headersJSON) == "" {
		return
	}
	var headers map[string]string
	if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
		fmt.Printf("[WARN] failed to parse custom headers: %v\n", err)
		return
	}
	for k, v := range headers {
		if IsBlockedHeader(k) {
			fmt.Printf("[WARN] blocked header %q skipped\n", k)
			continue
		}
		req.Header.Set(k, v)
	}
}

// GetWithHeaders 带自定义 Header 的 GET 请求
func GetWithHeaders(url string, headersJSON string, timeout int) ResponseWrapper {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return createRequestError(err)
	}
	SetCustomHeaders(req, headersJSON)
	return request(req, timeout)
}

// PostJsonWithHeaders 带自定义 Header 的 POST JSON 请求
func PostJsonWithHeaders(url string, body string, headersJSON string, timeout int) ResponseWrapper {
	buf := bytes.NewBufferString(body)
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return createRequestError(err)
	}
	req.Header.Set("Content-type", "application/json")
	SetCustomHeaders(req, headersJSON)
	return request(req, timeout)
}

// PostParamsWithHeaders 带自定义 Header 的 POST 表单请求
func PostParamsWithHeaders(url string, params string, headersJSON string, timeout int) ResponseWrapper {
	buf := bytes.NewBufferString(params)
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return createRequestError(err)
	}
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	SetCustomHeaders(req, headersJSON)
	return request(req, timeout)
}

func request(req *http.Request, timeout int) ResponseWrapper {
	wrapper := ResponseWrapper{StatusCode: 0, Body: "", Header: make(http.Header)}
	client := clientFactory(timeout)
	setRequestHeader(req)
	resp, err := client.Do(req)
	if err != nil {
		wrapper.Body = fmt.Sprintf("执行HTTP请求错误-%s", err.Error())
		return wrapper
	}
	defer resp.Body.Close()
	// 限制响应体最大 1MB，防止 OOM
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		wrapper.Body = fmt.Sprintf("读取HTTP请求返回值失败-%s", err.Error())
		return wrapper
	}
	wrapper.StatusCode = resp.StatusCode
	wrapper.Body = string(body)
	wrapper.Header = resp.Header

	return wrapper
}

func setRequestHeader(req *http.Request) {
	req.Header.Set("User-Agent", "golang/gocron")
}

func createRequestError(err error) ResponseWrapper {
	errorMessage := fmt.Sprintf("创建HTTP请求错误-%s", err.Error())
	return ResponseWrapper{0, errorMessage, make(http.Header)}
}
