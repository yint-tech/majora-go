package infra

import (
	"crypto/tls"
	"math/rand"
	"net/http"
	"time"

	"iinti.cn/majora-go/log"
)

// 网络检测

var (
	httpCli *http.Client
)

var (
	pingUrl = []string{
		"https://www.baidu.com",
		"https://www.bilibili.com",
		"https://www.taobao.com",
		"https://www.xiaohongshu.com",
		"https://www.bytedance.com",
		"https://pvp.qq.com",
	}
)

const (
	defTimeout = time.Second * 5
)

func init() {
	httpCli = &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout: defTimeout,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: defTimeout,
	}
}

func Ping(url string) bool {
	resp, err := httpCli.Head(url)
	if err != nil {
		log.Run().Warnf("ping %s with error %+v", url, err)
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode > 0
}

func RandomPing() bool {
	return Ping(RandUrl())
}

func RandUrl() string {
	rand.Seed(time.Now().UnixNano())
	return pingUrl[rand.Intn(len(pingUrl))]
}
