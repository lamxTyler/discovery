package consul

import (
	"os"

	"github.com/hashicorp/consul/api"
)

const (
	ENV_TAG_CONSUL_ADDR  = "CONSUL_ADDR"
	ENV_TAG_CONSUL_TOKEN = "CONSUL_TOKEN"
	ENV_TAG_CONSUL_DC    = "CONSUL_DC"
	ENV_TAG_CONSUL_ZONE  = "CONSUL_ZONE"
	ENV_TAG_CONSUL_ENV   = "CONSUL_ENV"
)

//NewClient new conusl api client by env
func NewClient() (*api.Client, error) {
	config := api.DefaultConfig()
	consulAddr := "127.0.0.1:8500" // dfault
	consulAddr, ok := os.LookupEnv(ENV_TAG_CONSUL_ADDR)
	if ok { //set by env
		config.Address = consulAddr
	}
	consulToken, ok := os.LookupEnv(ENV_TAG_CONSUL_TOKEN)
	if ok {
		config.Token = consulToken
	}
	consulDc, ok := os.LookupEnv(ENV_TAG_CONSUL_DC)
	if ok {
		config.Datacenter = consulDc
	}
	return api.NewClient(config)
}

//Zone 获取环境变量的区信息 区分相同服务的不同集群
func Zone() string {
	Zone, ok := os.LookupEnv(ENV_TAG_CONSUL_ZONE)
	if ok {
		return Zone
	}
	return "UNZONE"
}

//ENV 环境 TEST|PRODUCT|ALPHA|BETA
func ENV() string {
	environment, ok := os.LookupEnv(ENV_TAG_CONSUL_ENV)
	if ok {
		return environment
	}
	return "TEST"
}
