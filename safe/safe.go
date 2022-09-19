package safe

import "iinti.cn/majora-go/log"

func SafeGo(f func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Errorf("goroutine panic %+v", err)
			}
		}()
		f()
	}()
}
