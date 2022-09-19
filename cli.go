package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"iinti.cn/majora-go/infra"

	"iinti.cn/majora-go/client"
	"iinti.cn/majora-go/daemon"
	"iinti.cn/majora-go/global"
	"iinti.cn/majora-go/initialize"
	"iinti.cn/majora-go/log"
	"iinti.cn/majora-go/safe"
)

var (
	configure string
)

var (
	Version string
	Date    string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.StringVar(&configure, "conf", "", "./majora -c path/to/your/majora.ini")

	flag.Parse()
}

func initial() {
	if global.Config.PprofPort > 0 {
		safe.SafeGo(func() {
			addr := fmt.Sprintf("127.0.0.1:%d", global.Config.PprofPort)
			log.Run().Infof("enable pprof: %s", addr)
			log.Run().Error(http.ListenAndServe(addr, nil))
		})
	}

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
			log.Error().Errorf("cli panic %+v", err)
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

//main start
func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println(Version)
		os.Exit(0)
	}
	initialize.MustInitConfig(configure, global.Config)

	if global.Config.Daemon {
		logFile := filepath.Join(global.Config.LogPath, "daemon.log")
		d := daemon.NewDaemon(logFile)
		d.MaxCount = 20 //最大重启次数
		d.Run()
	}

	initialize.InitLogger()
	initial()
	cli()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-signalChan:
		time.Sleep(time.Second * 3)
		log.Run().Warn("main process exit...")
	}
}
