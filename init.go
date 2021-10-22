package discovery

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lamxTyler/discovery/consul"
	"github.com/lamxTyler/discovery/prometheus"
	"github.com/lamxTyler/discovery/utils/host"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ziipin-server/niuhe"
)

// PromHandler wrappers the standard http.Handler to gin.HandlerFunc
func PromHandler(handler http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
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
	return fmt.Sprintf("http://%s:%d", host.GetInnerIp(), port(serverAddr))
}

func handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"msg": "ok",
	})
}

func InitMonitoring(svr *niuhe.Server, pushIntervalSec int, pushAddr, serviceName string, watchPaths []string) {
	prometheus.InitMonitoring(svr, pushIntervalSec, pushAddr, serviceName, false, watchPaths)
	svr.GetGinEngine().GET("/metrics", PromHandler(promhttp.Handler()))
}

type Config struct {
	PushIntervalSec int
	PushAddr        string
	ServerAddr      string
	ServiceName     string
	Zone            string
	Idc             string
	ConsulAddr      string
	ConsulDc        string
	ConsulNode      string
	ConsulToken     string
	Debug           bool
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
	innerIP := host.GetInnerIp()
	if innerIP != "" {
		consul.Register(conf.ServiceName, accessAddr(conf.ServerAddr)+"/ping", map[string]string{}, accessAddr(conf.ServerAddr))
	}
}

func Init(svr *niuhe.Server, conf *Config, watchPaths []string) {
	InitMonitoring(svr, conf.PushIntervalSec, conf.PushAddr, conf.ServiceName, watchPaths)
	InitConsul(svr, conf)
}
