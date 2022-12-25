package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"iinti.cn/majora-go/infra" //nolint:gci

	"iinti.cn/majora-go/client"
	"iinti.cn/majora-go/daemon"
	"iinti.cn/majora-go/global" //nolint:gci
	"iinti.cn/majora-go/initialize"
	"iinti.cn/majora-go/log"
)

var configure string

var (
	Version string
	Date    string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.StringVar(&configure, "conf", "", "./majora -conf configure.yml")

	flag.Parse()
}

func initial() {
	if len(global.Config.DNSServer) > 0 {
		net.DefaultResolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return net.Dial("udp", global.Config.DNSServer)
			},
		}
	}
}

func cli() {
	defer func() {
		if err := recover(); err != nil {
			if err := recover(); err != nil {
				var buf [4096]byte
				n := runtime.Stack(buf[:], false)
				log.Error().Errorf("goroutine panic.stack:%s,err:%+v", string(buf[:n]), err)
			}
		}
	}()
	log.Run().Infof("cpu count %d proc %d", runtime.NumCPU(), runtime.NumCPU()*2)
	log.Run().Infof("current Version %s, build at %s", Version, Date)
	log.Run().Infof("hostInfo os:%s, arch:%s", runtime.GOOS, runtime.GOARCH)
	cfgInfo, _ := json.Marshal(global.Config)
	log.Run().Infof("config info:%s", string(cfgInfo))

	domainAndPort := strings.Split(global.Config.TunnelAddr, ":")
	if len(domainAndPort) != 2 {
		panic(errors.Errorf("TunnelAddr Error: %s", global.Config.TunnelAddr))
	}
	domain := domainAndPort[0]
	port, err := strconv.Atoi(domainAndPort[1])
	if err != nil {
		panic(errors.Errorf("Parse tunnel port error: %s", domainAndPort[1]))
	}
	clusterClient := client.ClusterClient{
		Domain: domain,
		Port:   port,
		Redial: infra.NewPPPRedial(),
	}
	clusterClient.Start()
}

// main start
func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println(Version)
		os.Exit(0)
	}
	initialize.MustInitConfig(configure, global.Config)

	if global.Config.Daemon {
		logFile := filepath.Join(global.Config.LogPath, "daemon.log")
		d := daemon.NewDaemon(logFile)
		d.MaxCount = 20 // 最大重启次数
		d.Run()
	}

	initialize.InitLogger()
	initial()
	cli()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	if <-signalChan; true {
		time.Sleep(time.Second * 3)
		log.Run().Warn("main process exit...")
	}
}
