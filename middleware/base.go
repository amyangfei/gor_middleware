package middleware

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

type GorMessage struct {
	Id      int
	Type    string
	Meta    [][]byte // Meta is an array size of 4, containing: request type, uuid, timestamp, latency
	RawMeta []byte   //
	Http    []byte   // Raw HTTP payload
}

type InterFunc struct {
	fn   func(*Gor, *GorMessage, ...interface{}) *GorMessage
	args []interface{}
}

type Gor struct {
	queue  map[string]([]*InterFunc)
	lock   *sync.RWMutex
	input  chan string
	stderr io.Writer
}

func CreateGor() *Gor {
	gor := &Gor{
		queue:  make(map[string]([]*InterFunc)),
		lock:   new(sync.RWMutex),
		input:  make(chan string),
		stderr: os.Stderr,
	}
	return gor
}

func (gor *Gor) On(
	channel string, fn func(*Gor, *GorMessage, ...interface{}) *GorMessage,
	idx string, args ...interface{}) {

	if idx != "" {
		channel = channel + "#" + idx
	}
	gor.lock.Lock()
	inmsg := &InterFunc{
		fn:   fn,
		args: args,
	}
	if c, ok := gor.queue[channel]; ok {
		c = append(c, inmsg)
	} else {
		newChan := make([]*InterFunc, 0)
		newChan = append(newChan, inmsg)
		gor.queue[channel] = newChan
	}
	gor.lock.Unlock()
}

func (gor *Gor) Emit(msg *GorMessage) error {
	chanPrefix, ok := ChanPrefixMap[msg.Type]
	if !ok {
		return errors.New(fmt.Sprintf("invalid message type: %s", msg.Type))
	}
	chanIds := [3]string{
		"message", chanPrefix, fmt.Sprintf("%s#%d", chanPrefix, msg.Id)}
	resp := msg
	for _, chanId := range chanIds {
		if funcs, ok := gor.queue[chanId]; ok {
			for _, f := range funcs {
				r := f.fn(gor, msg, f.args)
				if r != nil {
					resp = r
				}
			}
		}
	}
	if resp != nil {
		fmt.Printf("%s", gor.HexData(resp))
	}
	return nil
}

func (gor *Gor) HexData(msg *GorMessage) string {
	encodeList := [3][]byte{msg.RawMeta, []byte("\n"), msg.Http}
	encodedList := make([]string, 3)
	for i, val := range encodeList {
		encodedList[i] = hex.EncodeToString(val)
	}
	encodedList = append(encodedList, "\n")
	return strings.Join(encodedList, "")
}

func (gor *Gor) ParseMessage(line string) (*GorMessage, error) {
	payload, err := hex.DecodeString(strings.TrimSpace(line))
	if err != nil {
		return nil, err
	}
	metaPos := bytes.Index(payload, []byte("\n"))
	metaRaw := payload[:metaPos]
	metaArr := bytes.Split(metaRaw, []byte(" "))
	ptype := metaArr[0]
	pid, err := strconv.Atoi(string(metaArr[1]))
	if err != nil {
		return nil, err
	}
	httpPayload := payload[metaPos+1:]
	return &GorMessage{
		Id:      pid,
		Type:    string(ptype),
		Meta:    metaArr,
		RawMeta: metaRaw,
		Http:    httpPayload,
	}, nil
}

func (gor *Gor) preProcessor() {
	for {
		line := <-gor.input
		if msg, err := gor.ParseMessage(line); err != nil {
			gor.stderr.Write([]byte(err.Error()))
		} else {
			gor.Emit(msg)
		}
	}
}

func (gor *Gor) receiver() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		gor.input <- scanner.Text()
	}
}

func (gor *Gor) Run() {
	go gor.receiver()
	go gor.preProcessor()
}
