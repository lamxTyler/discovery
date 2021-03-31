package consul

import (
	"log"
	"os"
	"testing"
	"time"
)

const consulAddr = "127.0.0.1:8500"

// const consulAddr = "consul.badam.mobi:8500"

func TestRegist(t *testing.T) {
	// func Regist(servicesName, servicesAddr, httpPingAddr string)
	os.Setenv("CONSUL_ADDR", consulAddr)
	// os.Setenv("CONSUL_TOKEN", "8c3dc081-93ba-4f9b-b06b-dae7422548c3")
	Register("test-consul-discovery", DefaultHealthy, map[string]string{}, "ws://192.168.30.5:9090")
	// Deregister("test-consul-discovery", "ws://ght-hall.badambiz.com:9090")
	time.Sleep(time.Second * 60)
}

func TestDeregist(t *testing.T) {
	// func Regist(servicesName, servicesAddr, httpPingAddr string)
	os.Setenv("CONSUL_ADDR", consulAddr)
	os.Setenv("CONSUL_TOKEN", "8c3dc081-93ba-4f9b-b06b-dae7422548c3")
	// Deregister("test-consul-discovery", "ws://ght-hall.badambiz.com:9090")
	// Deregister("test-consul-discovery", "ws://ght-hall.badambiz.com:9090")
	deregisterByServiceID("test-consul-discovery#ght-hall.badambiz.com:")
}

func TestLookup(t *testing.T) {
	// func Regist(servicesName, servicesAddr, httpPingAddr string)
	os.Setenv("CONSUL_ADDR", consulAddr)
	// os.Setenv("CONSUL_TOKEN", "8c3dc081-93ba-4f9b-b06b-dae7422548c3")
	srvs, err := Services()
	if err != nil {
		panic(err)
	}
	for key, s := range srvs {
		log.Printf("serviceID:%s, s:%+v", key, s)
	}
}

func TestSvrAddrList(t *testing.T) {
	// func Regist(servicesName, servicesAddr, httpPingAddr string)
	os.Setenv("CONSUL_ADDR", consulAddr)
	// os.Setenv("CONSUL_TOKEN", "8c3dc081-93ba-4f9b-b06b-dae7422548c3")
	time.Sleep(time.Second)
	srvs := SvrAddrList("test-consul-discovery")
	for key, s := range srvs {
		log.Printf("serviceID:%d, s:%+v", key, s)
	}
}

func TestSvrAddrWithZone(t *testing.T) {
	// func Regist(servicesName, servicesAddr, httpPingAddr string)
	os.Setenv("CONSUL_ADDR", consulAddr)
	// os.Setenv("CONSUL_TOKEN", "8c3dc081-93ba-4f9b-b06b-dae7422548c3")
	time.Sleep(time.Second)
	srvs := SvrAddrWithZone("test-consul-discovery", "UNZONE")
	for key, s := range srvs {
		log.Printf("serviceID:%d, s:%+v", key, s)
	}
}

func TestSvrAddrWithTags(t *testing.T) {
	// func Regist(servicesName, servicesAddr, httpPingAddr string)
	os.Setenv("CONSUL_ADDR", consulAddr)
	// os.Setenv("CONSUL_TOKEN", "8c3dc081-93ba-4f9b-b06b-dae7422548c3")
	time.Sleep(time.Second)
	s, ok := Select(SvrAddrWithTags("test-consul-discovery", "UNZONE", "ws"))
	if !ok {
		panic("nofound")
	}
	log.Printf(" s:%+v", s)
}

func TestLookupByName(t *testing.T) {
	// func Regist(servicesName, servicesAddr, httpPingAddr string)
	os.Setenv("CONSUL_ADDR", consulAddr)
	// os.Setenv("CONSUL_TOKEN", "8c3dc081-93ba-4f9b-b06b-dae7422548c3")
	status, srvs, err := ServicesByName("test-consul-discovery")
	if err != nil {
		panic(err)
	}
	log.Printf("status:%s", status)
	for key, s := range srvs {
		// fmt.Println(key, "-", s)
		log.Printf("key:%d, s:%+v", key, s.Service)
	}
}

func TestCatalogServices(t *testing.T) {
	// func Regist(servicesName, servicesAddr, httpPingAddr string)
	os.Setenv("CONSUL_ADDR", consulAddr)
	os.Setenv("CONSUL_TOKEN", "8c3dc081-93ba-4f9b-b06b-dae7422548c3")
	srvs, err := CatalogServices()
	if err != nil {
		panic(err)
	}
	for key, tags := range srvs {
		// fmt.Println(key, "-", s)
		log.Printf("key:%s", key)
		for _, t := range tags {
			log.Printf("tag:%s", t)
		}
	}
}

func TestCatalogServicesByName(t *testing.T) {
	// func Regist(servicesName, servicesAddr, httpPingAddr string)
	os.Setenv("CONSUL_ADDR", consulAddr)
	os.Setenv("CONSUL_TOKEN", "8c3dc081-93ba-4f9b-b06b-dae7422548c3")
	srvs, err := CatalogServicesByName("test-consul-discovery", "")
	if err != nil {
		panic(err)
	}
	for key, s := range srvs {
		// fmt.Println(key, "-", s)
		log.Printf("key:%d , s: %+v", key, s)
		// for _, t := range tags {
		// 	log.Printf("tag:%s", t)
		// }
	}
}

//curl -X PUT -H "X-Consul-Token: 8c3dc081-93ba-4f9b-b06b-dae7422548c3" -d'{"Datacenter":"dc-master", "Node":"master", "ServiceID":"test-consul-discovery#ght-hall.badambiz.com:"}' http://consul.badam.mobi:8500/v1/catalog/deregister
//true%
