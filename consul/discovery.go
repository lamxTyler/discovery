package consul

import (
	"github.com/hashicorp/consul/api"
)

// ServicesByName returns for a given service name: the aggregated health status for all services
// having the specified name.
// - If no service is not found, will return status (critical, [], nil)
// - If the service is found, will return (critical|passing|warning), []api.AgentServiceChecksInfo, nil)
// - In all other cases, will return an error
func ServicesByName(serviceName string) (string, []api.AgentServiceChecksInfo, error) {
	consulCli, err := NewClient()
	if err != nil {
		return "", []api.AgentServiceChecksInfo{}, err
	}
	return consulCli.Agent().AgentHealthServiceByName(serviceName)
}

// Services returns the locally registered services
func Services() (map[string]*api.AgentService, error) {
	consulCli, err := NewClient()
	if err != nil {
		return nil, err
	}
	return consulCli.Agent().Services()
}

// ServicesWithFilter returns a subset of the locally registered services that match
// the given filter expression
func ServicesWithFilter(filter string) (map[string]*api.AgentService, error) {
	consulCli, err := NewClient()
	if err != nil {
		return nil, err
	}
	return consulCli.Agent().ServicesWithFilter(filter)
}

// CatalogServices is used to query for all known services
func CatalogServices() (map[string][]string, error) {
	consulCli, err := NewClient()
	if err != nil {
		return nil, err
	}
	lists, _, err := consulCli.Catalog().Services(&api.QueryOptions{})
	return lists, err
}

// CatalogServicesByName is used to query catalog entries for a given service
func CatalogServicesByName(serviceName, tag string) ([]*api.CatalogService, error) {
	consulCli, err := NewClient()
	if err != nil {
		return nil, err
	}
	services, _, err := consulCli.Catalog().Service(serviceName, tag, &api.QueryOptions{})
	return services, err
}
