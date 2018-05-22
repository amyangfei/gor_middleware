# GoReplay Middleware

[![Build Status](https://travis-ci.org/amyangfei/gor_middleware.svg?branch=master)](https://travis-ci.org/amyangfei/gor_middleware)

Golang library for [GoReplay Middleware](https://github.com/buger/goreplay) , API is quite similar to [NodeJS library](https://github.com/buger/goreplay/tree/master/middleware>)

## Installation

```bash
$ go get github.com/amyangfei/gor_middleware/gormw
```

## Getting Started

Sample code:

```go
package main

import (
        "fmt"
        "github.com/amyangfei/gor_middleware/gormw"
        "os"
)

func OnRequest(gor *gormw.Gor, msg *gormw.GorMessage, kwargs ...interface{}) *gormw.GorMessage {
        gor.On("response", OnResponse, msg.Id, msg)
        return msg
}

func OnResponse(gor *gormw.Gor, msg *gormw.GorMessage, kwargs ...interface{}) *gormw.GorMessage {
        req, _ := kwargs[0].(*gormw.GorMessage)
        gor.On("replay", OnReplay, req.Id, req, msg)
        return msg
}

func OnReplay(gor *gormw.Gor, msg *gormw.GorMessage, kwargs ...interface{}) *gormw.GorMessage {
        req, _ := kwargs[0].(*gormw.GorMessage)
        resp, _ := kwargs[1].(*gormw.GorMessage)
        fmt.Fprintf(os.Stderr, "request raw http: %s\n", req.Http)
        fmt.Fprintf(os.Stderr, "response raw http: %s\n", resp.Http)
        fmt.Fprintf(os.Stderr, "replay raw http: %s\n", msg.Http)
        respStatus, _ := gormw.HttpStatus(string(resp.Http))
        replayStatus, _ := gormw.HttpStatus(string(msg.Http))
        if respStatus != replayStatus {
                fmt.Fprintf(os.Stderr, "replay status [%s] diffs from response status [%s]\n", replayStatus, respStatus)
        } else {
                fmt.Fprintf(os.Stderr, "replay status is same as response status\n")
        }
        return msg
}

func main() {
        gor := gormw.CreateGor()
        gor.On("request", OnRequest, "")
        gor.Run()
}
```
