package client

type RedialCmd struct{}

func (r RedialCmd) Action() string {
	return ActionRedial
}

func (r RedialCmd) Handle(param map[string]string, callback Callback) {
	panic("implement me")
}

var redialCmd = &RedialCmd{}
