package httpclient

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockDoer func(req *http.Request) (*http.Response, error)

func (m mockDoer) Do(req *http.Request) (*http.Response, error) {
	return m(req)
}

func withMockClient(t *testing.T, doer mockDoer) {
	t.Helper()
	original := clientFactory
	clientFactory = func(timeout int) httpDoer {
		return doer
	}
	t.Cleanup(func() { clientFactory = original })
}

func TestGetRequest(t *testing.T) {
	withMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if ua := req.Header.Get("User-Agent"); ua != "golang/gocron" {
			t.Fatalf("unexpected user-agent %s", ua)
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("ok")),
			Header:     http.Header{},
		}, nil
	})

	resp := Get("http://example.com", 0)
	if resp.StatusCode != 200 || resp.Body != "ok" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestPostParamsRequest(t *testing.T) {
	withMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.Header.Get("Content-type") != "application/x-www-form-urlencoded" {
			t.Fatalf("unexpected content-type %s", req.Header.Get("Content-type"))
		}
		body, _ := io.ReadAll(req.Body)
		if string(body) != "a=1&b=2" {
			t.Fatalf("unexpected body %s", string(body))
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("echo:" + string(body))),
			Header:     http.Header{},
		}, nil
	})

	resp := PostParams("http://example.com", "a=1&b=2", 0)
	if resp.StatusCode != 200 || resp.Body != "echo:a=1&b=2" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestPostJsonRequest(t *testing.T) {
	withMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("Content-type") != "application/json" {
			t.Fatalf("unexpected content-type %s", req.Header.Get("Content-type"))
		}
		body, _ := io.ReadAll(req.Body)
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("json:" + string(body))),
			Header:     http.Header{},
		}, nil
	})

	resp := PostJson("http://example.com", `{"name":"gocron"}`, 0)
	if resp.StatusCode != 200 || resp.Body != `json:{"name":"gocron"}` {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestRequestHandlesClientError(t *testing.T) {
	withMockClient(t, func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("timeout")
	})
	resp := Get("http://example.com", 1)
	if resp.StatusCode != 0 || !strings.Contains(resp.Body, "执行HTTP请求错误-timeout") {
		t.Fatalf("expected client error message, got %+v", resp)
	}
}

func TestRequestHandlesReadError(t *testing.T) {
	withMockClient(t, func(req *http.Request) (*http.Response, error) {
		rc := io.NopCloser(io.Reader(&failingReader{}))
		return &http.Response{StatusCode: 200, Body: rc, Header: http.Header{}}, nil
	})
	resp := Get("http://example.com", 0)
	if resp.StatusCode != 0 || !strings.Contains(resp.Body, "读取HTTP请求返回值失败") {
		t.Fatalf("expected read error message, got %+v", resp)
	}
}

type failingReader struct{}

func (f *failingReader) Read(p []byte) (int, error) {
	return 0, errors.New("boom")
}

func TestCreateRequestError(t *testing.T) {
	resp := createRequestError(fmt.Errorf("boom"))
	if resp.StatusCode != 0 || !strings.Contains(resp.Body, "boom") {
		t.Fatalf("unexpected error wrapper: %+v", resp)
	}
}
