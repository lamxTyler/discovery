package consul

import (
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/lamxTyler/discovery/utils/destroy"
	"github.com/lamxTyler/discovery/utils/goroutine"
	"github.com/lamxTyler/discovery/utils/random"
	"github.com/patrickmn/go-cache"
	"github.com/ziipin-server/niuhe"
)

var registeredSvc map[string]bool

var defaultHTTPHealthyURL string

var agentServices *cache.Cache = cache.New(10*time.Second, 3*time.Second) //进程缓存值 有效期10s,检查间隔5s

//DefaultHealthy set healthy check by default
const DefaultHealthy = "default-healthy"

func markRegistered(serviceID string) {
	registeredSvc[serviceID] = true
}

func init() {
	registeredSvc = make(map[string]bool)
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "pong")
	})
	go func() {
		for {
			healthPort := strconv.Itoa(random.Between(30000, 40000))
			defaultHTTPHealthyURL = ":" + healthPort + "/ping"
			if err := http.ListenAndServe(":"+healthPort, nil); err != nil {
				niuhe.LogError("Error occur when start server %v", err)
				continue
			}
		}
	}()
	go destory()
	goroutine.Go(cachingServices)
}

func destory() {
	stopSignalChan := make(chan *sync.WaitGroup, 1)
	destroy.Register(stopSignalChan)
	for {
		select {
		case wg, ok := <-stopSignalChan:
			if ok {
				for serviceID := range registeredSvc {
					deregisterByServiceID(serviceID)
				}
				wg.Done()
			}
		}
	}
}

func cachingServices() {
	tick := time.NewTicker(time.Second * 5)
	defer tick.Stop()
	loadServices()
	for {
		select {
		case <-tick.C:
			if err := loadServices(); err != nil {
				continue
			}
		}
	}
}

func loadServices() error {
	list, err := Services()
	if err != nil {
		//niuhe.LogError("load services failed!! err:%v", err)
		return err
	}
	for _, s := range list {
		niuhe.LogDebug("add s:%+v", s)
		agentServices.SetDefault(s.ID, s)
	}
	return nil
}
