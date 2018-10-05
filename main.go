package main

import (
	"flag"
	"log"
	"sync"

	"github.com/safeie/goproxy/config"
	"github.com/safeie/goproxy/core"
	"github.com/safeie/goproxy/core/scheduler"
)

var lock sync.Mutex
var trueList []string
var protocol string
var local string
var server string

func main() {
	flag.StringVar(&protocol, "p", "tcp", "-p=tcp 指定协议类型")
	flag.StringVar(&local, "l", "", "-l=0.0.0.0:8080 指定本地监听端口")
	flag.StringVar(&server, "s", "", "-s=127.0.0.1:80,127.0.0.1:81 指定真实服务器地址,多个用','隔开")
	flag.Parse()

	config, err := config.New(protocol, local, server)
	if err != nil {
		log.Fatalln(err)
	}

	config.Debug = true
	config.Scheduler = scheduler.IPHashName
	config.Timeout = 2000
	proxy := core.New(config)
	proxy.Start()
}
