package initialize

import (
	"iinti.cn/majora-go/global"
	"iinti.cn/majora-go/log"
)

func InitLogger() {
	// 暂时在这里初始化环境
	_ = global.CurrentEnv.Set(global.Config.Env)
	log.Init(global.Config.LogLevel, global.Config.LogPath)
}
