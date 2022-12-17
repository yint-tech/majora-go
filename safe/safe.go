package safe

import (
	"runtime"

	"iinti.cn/majora-go/log"
)

func Go(name string, f func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				var buf [4096]byte
				n := runtime.Stack(buf[:], false)
				log.Error().Errorf("goroutine-[%s] panic.stack:%s,err:%+v", name, string(buf[:n]), err)
			}
		}()
		f()
	}()
}
