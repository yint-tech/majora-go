package client

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
	Handle(param map[string]string, callback Callback)
}

type CmdResponse struct {
	SerialNumber int64
	Client       *Client
}

func (c *CmdResponse) OnCmdResponse(_ bool, response map[string]string) {
	//packet := protocol.TypeControl.CreatePacket()
	//packet.SerialNumber = c.SerialNumber
	//packet.Data = protocol.EncodeExtra(response)
	//if err := c.Client.WriteAndFlush(packet); err != nil {
	//	logger.Error().Msgf("OnCmdResponse error %+v", err)
	//}
}

//func OnRedialCmdResponse(client *Client, serialNumber int64, success bool, response map[string]string) {
//	packet := protocol.TypeControl.CreatePacket()
//	packet.SerialNumber = serialNumber
//	packet.Data = protocol.EncodeExtra(response)
//	if err := client.WriteAndFlush(packet); err != nil {
//		logger.Error().Msgf("OnCmdResponse error %+v", err)
//	}
//	infra.Redial(client.config, client.cleanup)
//}
