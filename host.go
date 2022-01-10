package guangmu_go

import (
	"net"
	"os"
	"strings"
)

func GetHost() string {
	os.Hostname()
	host := os.Getenv("HOSTNAME")
	if host != "" {
		return host
	}
	return getInnerIp()
}

func getInnerIp() string {
	netInterfaces, err := net.Interfaces()
	if err != nil {
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
