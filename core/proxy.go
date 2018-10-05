package core

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/safeie/goproxy/config"
	"github.com/safeie/goproxy/core/server"
	"github.com/safeie/goproxy/core/tcp"
	"github.com/safeie/goproxy/core/udp"
)

// Proxy 代理实现
type Proxy struct {
	Config   *config.Config
	Shutdown chan struct{}
}

// New 创建新的代理实例
func New(config *config.Config) *Proxy {
	t := new(Proxy)
	t.Config = config
	t.Shutdown = make(chan struct{})
	return t
}

// Start 开始服务
func (t *Proxy) Start() {
	var s server.Server
	switch t.Config.Protocol {
	case "tcp":
		s = tcp.New(t.Config)
	case "udp":
		s = udp.New(t.Config)
	}
	go s.Start()
	go t.signal()

	<-t.Shutdown
	s.Stop()
	log.Println("proxy stopped")
}

// Stop 停止服务
func (t *Proxy) Stop() {
	t.Shutdown <- struct{}{}
}

// monitor 监听系统信号，重启或停止服务
func (t *Proxy) signal() {
	sch := make(chan os.Signal, 10)
	signal.Notify(sch, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT,
		syscall.SIGHUP, syscall.SIGSTOP, syscall.SIGQUIT)
	go func(ch <-chan os.Signal) {
		sig := <-ch
		log.Println("signal recieved " + sig.String() + ", at: " + time.Now().Format("2006-01-02 15:04:05"))
		t.Stop()
		if sig == syscall.SIGHUP {
			log.Println("proxy restart now...")
			procAttr := new(os.ProcAttr)
			procAttr.Files = []*os.File{nil, os.Stdout, os.Stderr}
			procAttr.Dir = os.Getenv("PWD")
			procAttr.Env = os.Environ()
			process, err := os.StartProcess(os.Args[0], os.Args, procAttr)
			if err != nil {
				log.Println("proxy restart process failed:" + err.Error())
				return
			}
			waitMsg, err := process.Wait()
			if err != nil {
				log.Println("proxy restart wait error:" + err.Error())
			}
			log.Println(waitMsg)
		} else {
			log.Println("proxy shutdown now...")
		}
	}(sch)
}
