package client

import (
	"iinti.cn/majora-go/infra"
	"iinti.cn/majora-go/safe"
)

type RedialCmd struct{}

func (r RedialCmd) Action() string {
	return ActionRedial
}

func (r RedialCmd) Handle(client *Client, param map[string]string, callback Callback) {
	rsp := make(map[string]string)
	if !infra.IsEnvOk() {
		rsp[KeyFailedMsg] = "redialEnv is not ok"
		rsp[KeyStatusCode] = "-1"
		callback.OnCmdResponse(false, rsp)
		return
	}
	rsp[KeyData] = "client accept redial task"
	rsp[KeyStatusCode] = "0"
	callback.OnCmdResponse(false, rsp)

	safe.Go("redial", func() {
		client.Redial("action")
		client.CloseAll()
		client.natTunnel.Close()
		client.connect()
	})
}

var redialCmd = &RedialCmd{}
