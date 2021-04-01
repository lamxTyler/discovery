package discovery

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lamxTyler/discovery/consul"
	"github.com/lamxTyler/discovery/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ziipin-server/niuhe"
)

// PromHandler wrappers the standard http.Handler to gin.HandlerFunc
func PromHandler(handler http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

func getInnerIp() string {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		niuhe.LogError("net.Interfaces failed, err:%v", err)
		return ""
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil && strings.HasPrefix(ipnet.IP.String(), "10.") {
						return ipnet.IP.String()
					}
				}
			}
		}
	}
	return ""
}

func port(serverAddr string) int {
	strs := strings.Split(serverAddr, ":")
	if len(strs) > 1 {
		port, _ := strconv.Atoi(strs[1])
		return port
	}
	port, _ := strconv.Atoi(strs[0])
	return port
}

func accessAddr(serverAddr string) string {
	return fmt.Sprintf("http://%s:%d", getInnerIp(), port(serverAddr))
}

func handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"msg": "ok",
	})
}

func InitMonitoring(svr *niuhe.Server, conf *Config, watchPaths []string) {
	options := prometheus.Options{
		ServiceName: conf.ServiceName,
		Idc:         conf.Idc,
		WatchPath:   make(map[string]struct{}),
		HistogramBuckets: []float64{
			0.5, 1, 3, 5, 10, 15, 20, 30, 40, 50, 75, 100, 150, 200, 400, 700, 1000, 2000, 3000, 5000, 10000,
		},
	}
	for _, watchPath := range watchPaths {
		options.WatchPath[watchPath] = struct{}{}
	}
	prometheus.InitCommonMonitoring(options)

	svr.Use(prometheus.PrometheusMiddlewareHandler())
	svr.GetGinEngine().GET("/metrics", PromHandler(promhttp.Handler()))
}

type Config struct {
	ServerAddr  string
	ServiceName string
	Zone        string
	Idc         string
	ConsulAddr  string
	ConsulDc    string
	ConsulNode  string
	ConsulToken string
	Debug       bool
}

func setConsulEnvs(conf *Config) {
	if conf.ConsulAddr != "" {
		os.Setenv(consul.ENV_TAG_CONSUL_ADDR, conf.ConsulAddr)
	}
	if conf.ConsulDc != "" {
		os.Setenv(consul.ENV_TAG_CONSUL_DC, conf.ConsulDc)
	}
	if conf.ConsulToken != "" {
		os.Setenv(consul.ENV_TAG_CONSUL_TOKEN, conf.ConsulToken)
	}
	if conf.Zone != "" {
		os.Setenv(consul.ENV_TAG_CONSUL_ZONE, conf.Zone)
	}
	if !conf.Debug {
		os.Setenv(consul.ENV_TAG_CONSUL_ENV, "PRODUCT")
	}
}

func InitConsul(svr *niuhe.Server, conf *Config) {
	setConsulEnvs(conf)
	svr.GetGinEngine().GET("/ping", handlePing)
	innerIP := getInnerIp()
	if innerIP != "" {
		consul.Register(conf.ServiceName, accessAddr(conf.ServerAddr)+"/ping", map[string]string{}, accessAddr(conf.ServerAddr))
	}
}

func Init(svr *niuhe.Server, conf *Config, watchPaths []string) {
	InitMonitoring(svr, conf, watchPaths)
	InitConsul(svr, conf)
}
