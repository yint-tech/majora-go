package initialize

import (
	"testing"

	"iinti.cn/majora-go/log"
)

func TestInitLogger(t *testing.T) {
	log.Init("debug", "")
	log.Run().Info("adfdfsfsf")
}
