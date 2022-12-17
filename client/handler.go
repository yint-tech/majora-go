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

func (c *CmdResponse) OnCmdResponse(_ bool, _ map[string]string) {
}
