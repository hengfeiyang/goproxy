package tcp

import (
	"io"
	"log"
	"net"
	"strings"

	"github.com/safeie/goproxy/config"
	"github.com/safeie/goproxy/core/scheduler"
	"github.com/safeie/goproxy/core/server"
)

// TCP proxy
type TCP struct {
	config   *config.Config
	listener net.Listener
}

// New return a new tcp proxy instance
func New(config *config.Config) *TCP {
	t := new(TCP)
	t.config = config
	return t
}

// Start listen and serve
func (t *TCP) Start() {
	var err error
	t.listener, err = net.Listen("tcp", t.config.Local)
	if err != nil {
		log.Printf("TCP.Start error: %v\n", err)
		return
	}
	defer t.Stop()
	log.Printf("TCP.Start %v, backends: %v\n", t.config.Local, t.config.Servers)

	for {
		conn, err := t.listener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), server.ErrNetClosing.Error()) {
				log.Printf("TCP.Listener.accept error: %v\n", err)
			}
			break
		}
		if t.config.Debug {
			log.Printf("TCP 客户端已连接: %v => %v\n", conn.RemoteAddr(), conn.LocalAddr())
		}
		go t.handle(conn)
	}
}

// Stop stop serve
func (t *TCP) Stop() {
	lic, ok := t.listener.(*net.TCPListener)
	if ok {
		lic.Close()
	}
}

func (t *TCP) handle(sconn net.Conn) {
	defer func() {
		log.Printf("TCP 客户端已关闭: %v => %v\n", sconn.RemoteAddr(), sconn.LocalAddr())
		sconn.Close()
	}()

	addr := scheduler.Get(t.config.Scheduler).Schedule(sconn.RemoteAddr().String(), t.config.Servers)
	dconn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("TCP 连接服务器 [%v] 失败: %v\n", addr, err)
		return
	}
	defer func() {
		log.Printf("TCP 服务端已关闭: %v => %v\n", dconn.LocalAddr(), addr)
		dconn.Close()
	}()
	log.Printf("TCP 服务端已连接: %v => %v\n", dconn.LocalAddr(), addr)

	closeChan := make(chan struct{}, 1)
	go func(sconn net.Conn, dconn net.Conn, closeChan chan struct{}) {
		_, err := io.Copy(dconn, sconn)
		if err != nil {
			if err == io.EOF {
				// 读取完毕
			} else if strings.Contains(err.Error(), server.ErrNetClosing.Error()) {
				// log.Printf("TCP 往 [%v] 发送数据失败: 连接已关闭连接 %v\n", addr, err)
			} else {
				log.Printf("TCP 往 [%v] 发送数据失败: %v\n", addr, err)
			}
		}
		closeChan <- struct{}{}
	}(sconn, dconn, closeChan)

	go func(sconn net.Conn, dconn net.Conn, closeChan chan struct{}) {
		_, err := io.Copy(sconn, dconn)
		if err != nil {
			if err == io.EOF {
				// 读取完毕
			} else if strings.Contains(err.Error(), server.ErrNetClosing.Error()) {
				// log.Printf("TCP 从 [%v] 接收数据失败: 连接已关闭连接 %v\n", ip, err)
			} else {
				log.Printf("TCP 从 [%v] 接收数据失败: %v\n", addr, err)
			}
		}
		closeChan <- struct{}{}
	}(sconn, dconn, closeChan)

	<-closeChan
}
