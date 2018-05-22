package gormw

import (
	"fmt"
	"github.com/amyangfei/gor_middleware/gormw"
	"reflect"
	"testing"
)

func TestHttpMethod(t *testing.T) {
	payload := "GET /test HTTP/1.1\r\n\r\n"
	if method, err := gormw.HttpMethod(payload); err != nil {
		t.Errorf(err.Error())
	} else if method != "GET" {
		t.Errorf("invalid method: %s", method)
	}
}

func TestHttpPath(t *testing.T) {
	payload := "GET /test HTTP/1.1\r\n\r\n"
	if path, err := gormw.HttpPath(payload); err != nil {
		t.Errorf(err.Error())
	} else if path != "/test" {
		t.Errorf("invalid path: %s", path)
	}

	if new_payload, err := gormw.SetHttpPath(payload, "/"); err != nil {
		t.Errorf(err.Error())
	} else if new_payload != "GET / HTTP/1.1\r\n\r\n" {
		t.Errorf("invalid new_payload: %s", new_payload)
	}

	if new_payload, err := gormw.SetHttpPath(payload, "/new/test"); err != nil {
		t.Errorf(err.Error())
	} else if new_payload != "GET /new/test HTTP/1.1\r\n\r\n" {
		t.Errorf("invalid new_payload: %s", new_payload)
	}
}

func TestHttpPathParam(t *testing.T) {
	payload := "GET /test HTTP/1.1\r\n\r\n"
	if params, err := gormw.HttpPathParam(payload, "test"); err != nil {
		t.Errorf(err.Error())
	} else if !reflect.DeepEqual(params, []string{}) {
		t.Errorf("invalid params: %v", params)
	}

	payload, err := gormw.SetHttpPathParam(payload, "test", "123")
	if err != nil {
		t.Errorf(err.Error())
	}
	if params, err := gormw.HttpPathParam(payload, "test"); err != nil {
		t.Errorf(err.Error())
	} else if !reflect.DeepEqual(params, []string{"123"}) {
		t.Errorf("invalid params: %v", params)
	}

	payload, err = gormw.SetHttpPathParam(payload, "qwer", "ty")
	if err != nil {
		t.Errorf(err.Error())
	}
	if params, err := gormw.HttpPathParam(payload, "qwer"); err != nil {
		t.Errorf(err.Error())
	} else if !reflect.DeepEqual(params, []string{"ty"}) {
		t.Errorf("invalid params: %v", params)
	}
}

func TestHttpHeader(t *testing.T) {
	payload := "GET / HTTP/1.1\r\nHost: localhost:3000\r\nUser-Agent: Golang\r\nContent-Length:5\r\n\r\nhello"
	expected := map[string]string{
		"host":           "localhost:3000",
		"User-Agent":     "Golang",
		"Content-Length": "5",
	}
	for name, value := range expected {
		header, err := gormw.HttpHeader(payload, name)
		if err != nil {
			t.Errorf(err.Error())
		}
		if header == nil || header["value"] != value {
			t.Errorf("invalid header %s value %s", name, header["value"])
		}
	}
}

func TestSetHttpHeader(t *testing.T) {
	payload := "GET / HTTP/1.1\r\nUser-Agent: Golang\r\nContent-Length: 5\r\n\r\nhello"
	uas := []string{"", "1", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_0)"}
	expected := "GET / HTTP/1.1\r\nUser-Agent: %s\r\nContent-Length: 5\r\n\r\nhello"
	for _, ua := range uas {
		new_payload, err := gormw.SetHttpHeader(payload, "user-agent", ua)
		if err != nil {
			t.Errorf(err.Error())
		}
		if new_payload != fmt.Sprintf(expected, ua) {
			t.Errorf("invalid payload after set http header: %s", new_payload)
		}
	}

	expected = "GET / HTTP/1.1\r\nX-Test: test\r\nUser-Agent: Golang\r\nContent-Length: 5\r\n\r\nhello"
	new_payload, err := gormw.SetHttpHeader(payload, "X-Test", "test")
	if err != nil {
		t.Errorf(err.Error())
	} else if new_payload != expected {
		t.Errorf("invalid payload after set http header: %s", new_payload)
	}

	expected = "GET / HTTP/1.1\r\nX-Test2: test2\r\nX-Test: test\r\nUser-Agent: Golang\r\nContent-Length: 5\r\n\r\nhello"
	new_payload, err = gormw.SetHttpHeader(new_payload, "X-Test2", "test2")
	if err != nil {
		t.Errorf(err.Error())
	} else if new_payload != expected {
		t.Errorf("invalid payload after set http header: %s", new_payload)
	}
}
