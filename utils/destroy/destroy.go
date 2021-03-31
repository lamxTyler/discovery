package destroy

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var registed = make([]chan *sync.WaitGroup, 0, 8)
var registerChan = make(chan chan *sync.WaitGroup)

//Register regist destroy event
func Register(p chan *sync.WaitGroup) {
	registerChan <- p
}

func watching() {
	stopSignalChan := make(chan os.Signal, 1)
	signal.Notify(stopSignalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	for {
		select {
		case c, ok := <-registerChan:
			if ok {
				registed = append(registed, c)
			}
		case _, ok := <-stopSignalChan:
			if ok {
				var wg = sync.WaitGroup{}
				wg.Add(len(registed))
				for _, c := range registed {
					c <- &wg
				}
				wg.Wait()
				os.Exit(0)
			}
		}
	}
}
func init() {
	go watching()
}
