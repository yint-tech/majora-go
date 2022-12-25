package client

import (
	"math/rand"
	"net"
	"sync"
	"time"

	"gopkg.in/fatih/set.v0"
	"iinti.cn/majora-go/global"
	"iinti.cn/majora-go/infra"
	"iinti.cn/majora-go/log"
	"iinti.cn/majora-go/safe"
)

type ClusterClient struct {
	Domain  string
	Port    int
	Redial  *infra.PPPRedial
	clients sync.Map
}

func (c *ClusterClient) Start() {
	c.check()
	c.RunStart()
	c.RunRedial()
}

func (c *ClusterClient) RunRedial() {
	if global.Config.Redial.Valid() {
		safe.Go("redial", func() {
			// 加上随机 防止vps在同时间重启
			duration := c.randomDuration()
			log.Run().Infof("Redial interval %+v", duration)
			timer := time.NewTimer(duration)
			for {
				<-timer.C
				c.StartRedial("cron", true)
				duration = c.randomDuration()
				log.Run().Infof("Redial interval %+v", duration)
				timer.Reset(duration)
			}
		})
	}
}

func (c *ClusterClient) RunStart() {
	safe.Go("start", func() {
		timer := time.NewTimer(5 * time.Minute)
		for {
			c.connectNatServers()
			<-timer.C
			timer.Reset(5 * time.Minute)
		}
	})
}

func (c *ClusterClient) connectNatServers() {
	hosts, err := net.LookupHost(c.Domain)
	if err != nil {
		log.Error().Errorf("[connectNatServers] lookup domain host error, %+v", err)
		return
	}
	log.Run().Infof("[connectNatServers] LookupHost from %s result: %+v", c.Domain, hosts)

	dnsSet := set.New(set.ThreadSafe)
	for _, v := range hosts {
		dnsSet.Add(v)
	}

	existSet := set.New(set.ThreadSafe)
	c.clients.Range(func(key, value interface{}) bool {
		existSet.Add(key)
		return true
	})

	needConnectSet := set.Difference(dnsSet, existSet)
	log.Run().Infof("[connectNatServers] NeedConnectSet: %v", needConnectSet.List())
	needConnectSet.Each(func(i interface{}) bool {
		t := i.(string)
		if _, ok := c.clients.Load(t); ok {
			log.Error().Error("[connectNatServers] client already exist")
		} else {
			client := NewClientWithConf(global.Config, t, c.Port)
			client.StartUp()
			c.clients.Store(t, client)
		}
		return true
	})

	needRemoveSet := set.Difference(existSet, dnsSet)
	log.Run().Infof("[connectNatServers] needRemoveSet: %v", needRemoveSet.List())
	needRemoveSet.Each(func(i interface{}) bool {
		t := i.(string)
		load, loaded := c.clients.LoadAndDelete(t)
		if !loaded {
			log.Error().Error("[connectNatServers] client already remove")
		}
		needCloseClient := load.(*Client)
		needCloseClient.CloseAll()
		needCloseClient.natTunnel.Close()
		return true
	})
}

func (c *ClusterClient) randomDuration() time.Duration {
	rand.Seed(time.Now().UnixNano())
	randDuration := rand.Int63n(time.Minute.Milliseconds() * 5)
	interval := randDuration + global.Config.Redial.RedialDuration.Milliseconds()
	return time.Duration(interval) * time.Millisecond
}

func (c *ClusterClient) StartRedial(tag string, replay bool) {
	defer func(startTime time.Time) {
		log.Run().Infof("StartRedial cost %v", time.Since(startTime))
	}(time.Now())
	if replay {
		c.clients.Range(func(host, c interface{}) bool {
			client, _ := c.(*Client)
			client.Redial(tag)
			return true
		})
		time.Sleep(global.Config.Redial.WaitTime)
	}
	c.Redial.Redial(tag)
	c.clients.Range(func(host, c interface{}) bool {
		client, _ := c.(*Client)
		client.CloseAll()
		client.natTunnel.Close()
		client.connect()
		return true
	})
}

func (c *ClusterClient) check() {
	if !global.Config.Redial.Valid() {
		return
	}
	interval := global.Config.NetCheckInterval
	if interval <= 0 {
		interval = time.Second * 5
	}

	url := infra.RandURL()
	if len(global.Config.NetCheckURL) > 0 {
		url = global.Config.NetCheckURL
	}

	safe.Go("check", func() {
		timer := time.NewTimer(interval)
		for {
			timer.Reset(interval)
			<-timer.C
			success := false
			for i := 0; i < 3; i++ {
				success = infra.Ping(url)
				if success {
					break
				}
			}
			if success {
				continue
			}
			log.Run().Warnf("Redial net check fail, redial...")
			c.StartRedial("check", false)
		}
	})
}
