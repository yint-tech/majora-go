package client

import (
	"github.com/adamweixuan/getty"
	"iinti.cn/majora-go/protocol"
)

var handlers = make(map[string]CmdHandler, 2)

func init() {
	handlers[shellCmd.Action()] = shellCmd
	handlers[redialCmd.Action()] = redialCmd
}

type CmdHandlerManager struct{}

func (CmdHandlerManager) HandleCmdMessage(client *Client, session getty.Session, packet *protocol.MajoraPacket) {
	param := protocol.DecodeExtra(packet.Data)
	action, ok := param[ACTION]

	hook := &CmdResponse{
		SerialNumber: packet.SerialNumber,
		Session:      session,
	}

	if !ok || len(action) == 0 {
		hook.OnCmdResponse(false, map[string]string{
			KeyFailedMsg:  "no param: {action} present",
			KeyStatusCode: "-1",
		})
		return
	}

	cmdHandler, ok := handlers[action]
	if !ok || cmdHandler == nil {
		hook.OnCmdResponse(false, map[string]string{
			KeyFailedMsg:  "no action: " + action + " defined",
			KeyStatusCode: "-1",
		})
		return
	}

	cmdHandler.Handle(client, param, hook)
}
