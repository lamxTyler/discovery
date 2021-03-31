package consul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/lamxTyler/discovery/utils/random"
)

//SvrAddrList 获取对应服务名的所有服务器
func SvrAddrList(svrName string) []string {
	svrMap := agentServices.Items()
	svrlist := make([]string, 0, len(svrMap))
	for _, item := range svrMap {
		if agentSvr, ok := item.Object.(*api.AgentService); ok {
			if agentSvr.Service == svrName {
				svrlist = append(svrlist, connectStr(agentSvr))
			}
		}
	}
	return svrlist
}

//SvrAddrWithZone 获取对应服务名以及大区的的所有服务器
func SvrAddrWithZone(svrName, zone string) []string {
	svrMap := agentServices.Items()
	svrlist := make([]string, 0, len(svrMap))
	for _, item := range svrMap {
		if agentSvr, ok := item.Object.(*api.AgentService); ok {
			if agentSvr.Service == svrName && agentSvr.Meta["zone"] == zone {
				svrlist = append(svrlist, connectStr(agentSvr))
			}
		}
	}
	return svrlist
}

//SvrAddrWithTags  获取对应服务名以及命中所有tag的的所有服务器
func SvrAddrWithTags(svrName string, tags ...string) []string {
	svrMap := agentServices.Items()
	svrlist := make([]string, 0, len(svrMap))
	for _, item := range svrMap {
		if agentSvr, ok := item.Object.(*api.AgentService); ok {
			if agentSvr.Service == svrName && checkTags(agentSvr.Tags, tags) {
				svrlist = append(svrlist, connectStr(agentSvr))
			}
		}
	}
	return svrlist
}

//Select 随机一个可用服务器
func Select(svrs []string) (string, bool) {
	if len(svrs) <= 0 {
		return "", false
	}
	return svrs[random.Intn(len(svrs))], true
}

func checkTags(tags, targetTags []string) bool {
	for _, tt := range targetTags {
		if !hasTags(tags, tt) {
			return false
		}
	}
	return true
}

func hasTags(tags []string, target string) bool {
	for _, tag := range tags {
		if tag == target {
			return true
		}
	}
	return false
}

func connectStr(svr *api.AgentService) string {
	return fmt.Sprintf("%s:%d", svr.Address, svr.Port)
}
