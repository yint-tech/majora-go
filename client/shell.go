package client

import (
	"os/exec"
	"strings"

	"iinti.cn/majora-go/log"
)

var cmdErrorMap = map[string]string{
	KeyFailedMsg:  "no param:{cmd} present",
	KeyStatusCode: "-1",
}

var shellCmd = &ShellCmd{}

type ShellCmd struct{}

func (e *ShellCmd) Action() string {
	return ActionExecShell
}

func (e *ShellCmd) Handle(param map[string]string, callback Callback) {
	targetCmd := param["cmd"]
	if len(targetCmd) == 0 || len(strings.TrimSpace(targetCmd)) == 0 {
		callback.OnCmdResponse(false, cmdErrorMap)
		return
	}
	log.Run().Infof("exec cmd %s", targetCmd)

	trueCmd := strings.Split(targetCmd, " ")

	cmd := exec.Command(trueCmd[0], trueCmd[1:]...) //nolint:gosec
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Run().Errorf("exec error %+v", err)
		return
	}

	callback.OnCmdResponse(true, map[string]string{
		KeyData:       string(out),
		KeyStatusCode: "0",
	})
}
