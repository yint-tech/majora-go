package infra

import (
	"os"
	"os/exec"
	"runtime"
	"time"

	"go.uber.org/atomic"
	"iinti.cn/majora-go/global" //nolint:gci
	"iinti.cn/majora-go/log"
)

const (
	cmdWin  = "/C"
	cmdUnix = "-c"
)

type PPPRedial struct {
	inRedialing *atomic.Bool
}

func NewPPPRedial() *PPPRedial {
	return &PPPRedial{
		inRedialing: atomic.NewBool(false),
	}
}

func (p *PPPRedial) Redial(tag string) bool {
	if p.inRedialing.CompareAndSwap(false, true) {
		log.Run().Infof("[PPPRedial %s] start", tag)
		beforeIP := GetPPP()
		retry := 0
		defer func(start time.Time) {
			newIP := GetPPP()
			log.Run().Infof("[PPPRedial %s] retry %d, cost %v, ip change %s -> %s ",
				tag, retry, time.Since(start), beforeIP, newIP)
		}(time.Now())
		for {
			retry++
			status := Command()
			pingBaidu := RandomPing()
			log.Run().Infof("[PPPRedial %s] net check: %d->%v", tag, retry, pingBaidu)
			if pingBaidu && status {
				break
			}
		}
		p.inRedialing.CompareAndSwap(true, false)
		return true
	}
	log.Run().Infof("[PPPRedial %s] inRedialing ignore this", tag)
	return false
}

func (p *PPPRedial) RedialByCheck() bool {
	return p.Redial("check")
}

func IsEnvOk() bool {
	execPath := global.Config.Redial.ExecPath
	if len(execPath) == 0 {
		log.Run().Warn("[Redial] exec file is empty")
		return false
	}

	if _, err := os.Stat(execPath); err != nil {
		log.Run().Warn("[Redial] stat execpath error:%s", err.Error())
		return false
	}

	command := global.Config.Redial.Command
	if len(command) == 0 {
		log.Run().Warn("[Redial] Command is empty")
		return false
	}

	if _, err := os.Stat(command); err != nil {
		log.Run().Warn("[Redial] stat command error:%s", err.Error())
		return false
	}
	return true
}

func Command() bool {
	if !IsEnvOk() {
		return true
	}
	execPath := global.Config.Redial.ExecPath
	command := global.Config.Redial.Command

	args := cmdUnix
	if runtime.GOOS == "windows" {
		args = cmdWin
	}

	cmd := exec.Command(command, args, execPath)
	output, err := cmd.Output()
	if err != nil {
		log.Run().Errorf("[Redial] Execute Shell:%s failed with error:%s", command, err.Error())
		return false
	}
	log.Run().Infof("[Redial] success %+v resp:%s", cmd, string(output))
	return true
}
