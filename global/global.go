package global //nolint:typecheck

import (
	"iinti.cn/majora-go/env"
	"iinti.cn/majora-go/model"
)

var (
	Config     = model.NewDefMajoraConf()
	CurrentEnv = env.Product
)
