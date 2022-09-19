package initialize

import (
	"fmt"
	"testing"
	"iinti.cn/majora-go/global"
)

func TestMustInitConfig(t *testing.T) {
	MustInitConfig("/Users/tsaiilin/src/go/majora-go/conf/majora-dev.yaml", global.Config)
	fmt.Printf("%+v", global.Config)

}
