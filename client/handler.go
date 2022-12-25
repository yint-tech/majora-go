package client

import (
	"github.com/adamweixuan/getty"
	"iinti.cn/majora-go/log"
	"iinti.cn/majora-go/protocol"
)

const (
	ActionExecShell = "executeShell"
	ActionRedial    = "redial"
	ACTION          = "action"
	KeyFailedMsg    = "errorMsg"
	KeyStatusCode   = "status"
	KeyData         = "data"
)

type Callback interface {
	OnCmdResponse(bool, map[string]string)
}

type CmdHandler interface {
	Action() string
	Handle(client *Client, param map[string]string, callback Callback)
}

type CmdResponse struct {
	SerialNumber int64
	Session      getty.Session
}

func (c *CmdResponse) OnCmdResponse(_ bool, response map[string]string) {
	packet := protocol.TypeControl.CreatePacket()
	packet.SerialNumber = c.SerialNumber
	packet.Extra = string(protocol.EncodeExtra(response))
	log.Run().Info("OnCmdResponse run:%+v", response)
	_, _, err := c.Session.WritePkg(packet, 0) //nolint:errcheck
	if err != nil {
		log.Run().Error("OnCmdResponse run error:%s", err.Error())
	}
}
