package infra

import (
	"os"
	"os/exec"
	"time"

	"iinti.cn/majora-go/log"
)

func Restart() {
	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	log.Run().Debugf("Restart ... %+v", cmd)

	if err := cmd.Run(); err != nil {
		log.Run().Errorf("Restart error %+v", err)
	}

}

func RestartBySignal(signal chan struct{}) {
	go func() {
		time.Sleep(time.Second * 5)
		signal <- struct{}{}
	}()

	cmd := exec.Command(os.Args[0], os.Args[1:]...) //nolint:gosec
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	log.Run().Infof("[RestartBySignal] ... %+v", cmd)
	if err := cmd.Run(); err != nil {
		log.Run().Errorf("Restart error %+v", err)
	}
}
