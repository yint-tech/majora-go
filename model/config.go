package model

import (
	"os"
	"time"

	"github.com/google/uuid"
	"iinti.cn/majora-go/common"
	"iinti.cn/majora-go/log"
)

type Redial struct {
	Command        string        `mapstructure:"command"`
	ExecPath       string        `mapstructure:"exec_path"`
	RedialDuration time.Duration `mapstructure:"redial_duration"`
	WaitTime       time.Duration `mapstructure:"wait_time"`
}

type Extra struct {
	Account string `mapstructure:"account"`
}

type Configure struct {
	Env              string        `mapstructure:"env"`
	LogLevel         string        `mapstructure:"log_level"`
	LogPath          string        `mapstructure:"log_path"`
	Daemon           bool          `mapstructure:"daemon"`
	PprofPort        int           `mapstructure:"pprof_port"`
	TunnelAddr       string        `mapstructure:"tunnel_addr"`
	DNSServer        string        `mapstructure:"dns_server"`
	LocalAddr        string        `mapstructure:"local_ip"`
	ReconnInterval   time.Duration `mapstructure:"reconn_interval"`
	ClientID         string        `mapstructure:"client_id"`
	NetCheckInterval time.Duration `mapstructure:"net_check_interval"`
	NetCheckURL      string        `mapstructure:"net_check_url"`
	DNSCacheDuration time.Duration `mapstructure:"dns_cache_duration"`
	Extra            Extra         `mapstructure:"extra"`
	Redial           Redial        `mapstructure:"redial"`
}

const (
	reconninterval = time.Second * 10
)

func NewDefMajoraConf() *Configure {
	return &Configure{
		Env:            "product",
		LogLevel:       "info",
		Daemon:         false,
		PprofPort:      0,
		TunnelAddr:     common.DefNatAddr,
		DNSServer:      common.DNSServer, //nolint:typecheck
		ReconnInterval: reconninterval,
		ClientID:       uuid.New().String(),
		Redial: Redial{
			RedialDuration: reconninterval,
		},
	}
}

func (r Redial) Valid() bool {
	if len(r.Command) == 0 {
		return false
	}

	if len(r.ExecPath) == 0 {
		return false
	}

	if _, err := os.Stat(r.ExecPath); err != nil {
		log.Run().Errorf("read %s with error %s", r.ExecPath, err.Error())
		return false
	}

	if r.RedialDuration <= 0 {
		return false
	}
	return true
}
