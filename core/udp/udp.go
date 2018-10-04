package udp

import (
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/safeie/goproxy/config"
	"github.com/safeie/goproxy/core/scheduler"
	"github.com/safeie/goproxy/core/server"
)

// UDP proxy
type UDP struct {
	config        *config.Config
	listener      *net.UDPConn
	connStore     *sync.Map
	channelServer chan message
	channelClient chan message
}

// udp message
type message struct {
	Data []byte
	Conn *net.UDPConn
	Addr *net.UDPAddr
}

type conn struct {
	Conn   *net.UDPConn
	Active time.Time
}

// New return a new udp proxy instance
func New(config *config.Config) *UDP {
	t := new(UDP)
	t.config = config
	t.connStore = new(sync.Map)
	t.channelServer = make(chan message, 1024)
	t.channelClient = make(chan message, 1024)
	return t
}

// Start listen and serve
func (t *UDP) Start() {
	addr, err := net.ResolveUDPAddr("udp", t.config.Local)
	if err != nil {
		log.Printf("UDP.Start error: %v\n", err)
		return
	}
	t.listener, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Printf("UDP.Start error: %v\n", err)
		return
	}
	defer t.Stop()
	log.Printf("UDP.Start %v, backends: %v\n", t.config.Local, t.config.Servers)

	go t.handleServer()
	go t.handleClient()
	go t.gc()

	for {
		data := make([]byte, 4096)
		n, remoteAddr, err := t.listener.ReadFromUDP(data)
		if err != nil {
			if !strings.Contains(err.Error(), server.ErrNetClosing.Error()) {
				log.Printf("UDP.listener.Read error: %v\n", err)
			}
			break
		}
		if t.config.Debug {
			_, ok := t.connStore.Load(remoteAddr.String())
			if !ok {
				log.Printf("UDP 客户端已连接: %v => %v\n", remoteAddr.String(), t.listener.LocalAddr())
			}
		}
		t.channelServer <- message{
			Data: data[:n],
			Conn: t.listener,
			Addr: remoteAddr,
		}
	}
}

// Stop stop serve
func (t *UDP) Stop() {
	close(t.channelServer)
	close(t.channelClient)
	t.listener.Close()
}

func (t *UDP) handleServer() {
	for msg := range t.channelServer {
		var dconn *net.UDPConn
		c, ok := t.connStore.Load(msg.Addr.String())
		if !ok {
			serAddr := scheduler.Get(t.config.Scheduler).Schedule(msg.Addr.String(), t.config.Servers)
			udpAddr, err := net.ResolveUDPAddr("udp", serAddr)
			dconn, err = net.DialUDP("udp", nil, udpAddr)
			if err != nil {
				log.Printf("UDP 连接服务器 [%v] 失败: %v\n", serAddr, err)
				break
			}
			t.connStore.Store(msg.Addr.String(), &conn{Conn: dconn, Active: time.Now()})
			log.Printf("UDP 服务端已连接: %v => %v\n", dconn.LocalAddr(), serAddr)
		} else {
			dconn = c.(*conn).Conn
		}

		dconn.Write(msg.Data)
		t.connStore.Store(msg.Addr.String(), &conn{Conn: dconn, Active: time.Now()})
		go func(msg message) {
			for {
				data := make([]byte, 4096)
				n, _, err := dconn.ReadFromUDP(data)
				if err != nil {
					if !strings.Contains(err.Error(), server.ErrNetClosing.Error()) {
						log.Printf("UDP 从 [%v] 接收数据失败: %v\n", dconn.RemoteAddr().String(), err)
					}
					break
				}
				t.channelClient <- message{
					Data: data[:n],
					Conn: msg.Conn,
					Addr: msg.Addr,
				}
			}
		}(msg)
	}
}

func (t *UDP) handleClient() {
	for msg := range t.channelClient {
		_, err := msg.Conn.WriteToUDP(msg.Data, msg.Addr)
		if err != nil {
			log.Printf("UDP 往 [%v] 发送数据失败: %v\n", msg.Addr.String(), err)
		}
	}
}

func (t *UDP) gc() {
	for {
		time.Sleep(time.Second * 10)
		t.connStore.Range(func(key, val interface{}) bool {
			conn, ok := val.(*conn)
			if !ok {
				t.connStore.Delete(key)
				return false
			}
			if conn.Active.Add(time.Second * 30).Before(time.Now()) {
				if t.config.Debug {
					log.Printf("UDP 服务端已释放: %v => %v\n", conn.Conn.LocalAddr().String(), conn.Conn.RemoteAddr().String())
				}
				conn.Conn.Close()
				t.connStore.Delete(key)
			}
			return true
		})
	}
}
