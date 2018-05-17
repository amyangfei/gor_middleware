package middleware

import (
	"encoding/hex"
	"testing"
)

var passby map[string]int = make(map[string]int)

func init() {
	passby["received"] = 0
}

func incrReceived(gor *Gor, msg *GorMessage, kwargs ...interface{}) *GorMessage {
	passbyReadonly, _ := kwargs[0].(map[string]int)
	passby["received"] += passbyReadonly["received"] + 1
	return msg
}

func TestMessageLogic(t *testing.T) {
	gor := CreateGor()
	gor.On("message", incrReceived, "", &passby)
	gor.On("request", incrReceived, "", &passby)
	gor.On("response", incrReceived, "2", &passby)
	if len(gor.retainQueue) != 2 {
		t.Errorf("gor retain queue length %d != 2", len(gor.retainQueue))
	}
	if len(gor.tempQueue) != 1 {
		t.Errorf("gor temp queue length %d != 1", len(gor.tempQueue))
	}
	req, err := gor.ParseMessage(hex.EncodeToString([]byte("1 2 3\nGET / HTTP/1.1\r\n\r\n")))
	if err != nil {
		t.Error(err.Error())
	}
	resp, err := gor.ParseMessage(hex.EncodeToString([]byte("2 2 3\nHTTP/1.1 200 OK\r\n\r\n")))
	if err != nil {
		t.Error(err.Error())
	}
	resp2, err := gor.ParseMessage(hex.EncodeToString([]byte("2 3 3\nHTTP/1.1 200 OK\r\n\r\n")))
	if err != nil {
		t.Error(err.Error())
	}
	gor.Emit(req)
	gor.Emit(resp)
	gor.Emit(resp2)
	if passby["received"] != 5 {
		t.Errorf("passby received %d != 5", passby["received"])
	}
}
