package consul

import (
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/hashicorp/consul/api"
	"github.com/ziipin-server/niuhe"
)

//Register services to consul by default args
func Register(servicesName, httpHealthyAddr string, metas map[string]string, servicesAddrs ...string) error {
	consulCli, err := NewClient()
	if err != nil {
		log.Fatal("register server error : ", err)
	}
	for _, servicesAddr := range servicesAddrs {
		err = consulCli.Agent().ServiceRegister(genAgentServiceRegistration(servicesName, servicesAddr, httpHealthyAddr, metas))
		if err != nil {
			niuhe.LogError("regist services:%s servicesAddr:%s err:%v", servicesName, servicesAddr, err)
			return err
		}
		markRegistered(serivcesID(servicesName, servicesAddr))
	}
	return nil
}

//Deregister services to consul by default args
func Deregister(servicesName, servicesAddr string) error {
	return deregisterByServiceID(serivcesID(servicesName, servicesAddr))
}

func deregisterByServiceID(serivcesID string) error {
	consulCli, err := NewClient()
	if err != nil {
		niuhe.LogError("register server error : %+v ", err)
		return err
	}
	err = consulCli.Agent().ServiceDeregister(serivcesID)
	if err != nil {
		niuhe.LogError("deregist servicesid:%s err:%v", serivcesID, err)
		return err
	}
	return nil
}

func convert(servicesAddr string) *url.URL {
	urlObj, err := url.Parse(servicesAddr)
	if err != nil {
		panic(err)
	}
	return urlObj
}
func port(servicesAddr string) int {
	obj := convert(servicesAddr)
	port, _ := strconv.Atoi(obj.Port())
	return port
}
func host(servicesAddr string) string {
	obj := convert(servicesAddr)
	return obj.Hostname()
}

func serivcesID(servicesName, servicesAddr string) string {
	urlobj := convert(servicesAddr)
	return fmt.Sprintf("%s#%s:%s", servicesName, urlobj.Hostname(), urlobj.Port())
}

func checkID(servicesName, servicesAddr string) string {
	// urlobj := convert(servicesAddr)
	// return fmt.Sprintf("%s#%s:%s-check", servicesName, urlobj.Host, urlobj.Port())
	return serivcesID(servicesName, servicesAddr) + "#check"
}

func genAgentServiceRegistration(servicesName, servicesAddr, httpPingAddr string, metas map[string]string) *api.AgentServiceRegistration {
	urlobj := convert(servicesAddr)
	if httpPingAddr == DefaultHealthy {
		httpPingAddr = "http://" + urlobj.Hostname() + defaultHTTPHealthyURL
	}
	_metas := make(map[string]string)
	_metas["scheme"] = urlobj.Scheme
	_metas["zone"] = Zone()
	_metas["evn"] = ENV()
	for key, val := range metas {
		_metas[key] = val
	}
	_tags := make([]string, 0, len(_metas))
	for _, val := range _metas {
		_tags = append(_tags, val)
	}
	return &api.AgentServiceRegistration{
		Kind:    api.ServiceKindTypical,
		ID:      serivcesID(servicesName, servicesAddr),
		Name:    servicesName,
		Tags:    _tags,
		Meta:    _metas,
		Port:    port(servicesAddr),
		Address: urlobj.Hostname(),
		Check: &api.AgentServiceCheck{
			CheckID:                        checkID(servicesName, servicesAddr),
			Name:                           servicesName + "-check",
			Status:                         "passing",
			Interval:                       "5s",
			Timeout:                        "3s",
			HTTP:                           httpPingAddr,
			DeregisterCriticalServiceAfter: "30s",
		},
	}
}
