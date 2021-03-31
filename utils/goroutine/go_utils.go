package goroutine

import (
	// "common/zenv"
	"os"
	"runtime/debug"

	"github.com/ziipin-server/niuhe"
)

func CommonDefer() {
	if err := recover(); err != nil {
		niuhe.LogFatal("goroutine panic() err:%v", err)
		niuhe.LogFatal("%s", string(debug.Stack()))
		os.Exit(-1)
	}
}

// 公共关键协程使用GO 崩溃时用 supervisor 把整个进程重新拉起来
func Go(proc func()) {
	go func() {
		defer CommonDefer()
		proc()
	}()
}

func HandleGo(fn, handleFn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				niuhe.LogFatal("goroutine panic() err:%v", err)
				niuhe.LogFatal("%s", string(debug.Stack()))
				handleFn()
			}
		}()
		fn()
	}()
}
