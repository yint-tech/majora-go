package client

import (
	"net"
	"sync"

	"github.com/adamweixuan/getty"
	"github.com/coocood/freecache"

	"iinti.cn/majora-go/model"
)

type Client struct {
	config *model.Configure

	host         string
	port         int
	localAddr    net.Addr
	natTunnel    getty.Client
	session      getty.Session
	connStore    sync.Map
	sessionStore sync.Map
	dnsCache     *freecache.Cache
}

func NewClientWithConf(cfg *model.Configure, host string, port int) *Client {
	return NewCli(cfg, host, port)
}

func NewCli(cfg *model.Configure, host string, port int) *Client {
	var localAddr net.Addr
	if len(cfg.LocalAddr) > 0 {
		localAddr = &net.TCPAddr{
			IP:   net.ParseIP(cfg.LocalAddr),
			Port: 0,
		}
	}
	client := &Client{
		config:       cfg,
		host:         host,
		port:         port,
		localAddr:    localAddr,
		connStore:    sync.Map{},
		sessionStore: sync.Map{},
		dnsCache:     freecache.NewCache(1024),
	}

	return client
}

func (client *Client) StartUp() {
	client.connect()
}
