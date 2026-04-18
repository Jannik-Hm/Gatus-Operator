package annotatedressources

import (
	"context"
	"fmt"
	"slices"

	"github.com/Jannik-Hm/Gatus-Operator/internal/config"
	gatusconfig "github.com/Jannik-Hm/Gatus-Operator/internal/gatus_config"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type AnnotatedHTTPRoute struct {
	*gatewayv1.HTTPRoute
	gateways GatewayMap
}

func NewAnnotatedHTTPRoute(ctx context.Context, client client.Reader, route *gatewayv1.HTTPRoute) (*AnnotatedHTTPRoute, error) {
	annotated_route := &AnnotatedHTTPRoute{
		HTTPRoute: route,
		gateways:  GatewayMap{},
	}

	err := annotated_route.GetMissingRouteGateways(ctx, client)
	if err != nil {
		return nil, err
	}
	return annotated_route, nil
}

// key consists of namespace/name
type GatewayMap map[string]*gatewayv1.Gateway

func (obj *AnnotatedHTTPRoute) GetURLs() ([]string, error) {
	urls := make([]string, 0)

	paths := obj.getPaths()

	for host, protocols := range obj.getUrlsAndProtocols() {
		for protocol_type, _ := range protocols {
			protocol, err := getGatewayProtocolType(protocol_type)
			if err != nil {
				return nil, fmt.Errorf("Error generating URLs: %s", err)
			}
			for host_path, _ := range paths {
				urls = append(urls, fmt.Sprintf("%s://%s%s", protocol, host, host_path))
			}
		}
	}

	return urls, nil
}

func (obj *AnnotatedHTTPRoute) GetConditions(protocol string) []string {
	switch protocol {
	case "http", "https":
		return []string{"[STATUS] == 200"}
	case "tcp", "udp":
		return []string{"[CONNECTED] == true"}
	default:
		return []string{fmt.Sprintf("unknown type %s", protocol)}
	}
}

func (obj *AnnotatedHTTPRoute) getHostnames() []string {
	hostnames := make([]string, 0, len(obj.Spec.Hostnames))
	for _, hostname := range obj.Spec.Hostnames {
		hostnames = append(hostnames, string(hostname))
	}
	return hostnames
}

func (obj *AnnotatedHTTPRoute) getPaths() map[string]struct{} {
	paths := map[string]struct{}{}
	for _, rule := range obj.Spec.Rules {
		for _, path := range rule.Matches {
			if path.Path.Type == nil ||
				(*path.Path.Type != gatewayv1.PathMatchExact && *path.Path.Type != gatewayv1.PathMatchPathPrefix) {
				continue
			}
			paths[*path.Path.Value] = struct{}{}
		}
	}
	return paths
}

type protocolSet map[gatewayv1.ProtocolType]struct{}
type domain string

func (obj *AnnotatedHTTPRoute) getUrlsAndProtocols() map[domain]protocolSet {
	protocols := map[domain]protocolSet{}

	urls := obj.getHostnames()

	for _, parent_ref := range obj.Spec.ParentRefs {
		gateway, listener_name := obj.getGateway(parent_ref)
		if gateway == nil {
			continue
		}

		for _, listener := range gateway.Spec.Listeners {
			listener_urls := urls

			if listener.Hostname != nil && (len(urls) == 0 || slices.Contains(listener_urls, string(*listener.Hostname))) {
				listener_urls = []string{string(*listener.Hostname)}
			}

			if listener_name == nil || listener.Name == *listener_name {
				for _, url := range listener_urls {
					if _, ok := protocols[domain(url)]; !ok {
						protocols[domain(url)] = protocolSet{}
					}
					protocols[domain(url)][listener.Protocol] = struct{}{}
				}
			}
		}
	}

	return protocols
}

func (obj *AnnotatedHTTPRoute) addGateway(gateway *gatewayv1.Gateway) {
	if gateway == nil {
		return
	}

	gateway_key := fmt.Sprintf("%s/%s", gateway.GetNamespace(), gateway.GetName())

	if _, ok := obj.gateways[gateway_key]; !ok {
		obj.gateways[gateway_key] = gateway
	}
}

func (obj *AnnotatedHTTPRoute) getGateway(parent_ref gatewayv1.ParentReference) (*gatewayv1.Gateway, *gatewayv1.SectionName) {
	if parent_ref.Kind != nil && *parent_ref.Kind != "Gateway" {
		return nil, nil
	}

	listener_name := parent_ref.SectionName

	var gateway_namespace string = obj.GetNamespace()
	if parent_ref.Namespace != nil {
		gateway_namespace = string(*parent_ref.Namespace)
	}

	gateway_key := fmt.Sprintf("%s/%s", gateway_namespace, parent_ref.Name)
	return obj.gateways[gateway_key], listener_name
}

func (obj *AnnotatedHTTPRoute) GetMissingRouteGateways(ctx context.Context, client client.Reader) error {
	for _, parent_ref := range obj.Spec.ParentRefs {
		if parent_ref.Kind != nil && *parent_ref.Kind != "Gateway" {
			continue
		}

		var gateway_namespace string = obj.GetNamespace()
		if parent_ref.Namespace != nil {
			gateway_namespace = string(*parent_ref.Namespace)
		}

		gateway_ressource_key := types.NamespacedName{
			Name:      string(parent_ref.Name),
			Namespace: gateway_namespace,
		}

		gateway := &gatewayv1.Gateway{}

		if err := client.Get(ctx, gateway_ressource_key, gateway); err != nil {
			return fmt.Errorf("Could not get Gateway: %s", err)
		}

		obj.addGateway(gateway)
	}

	return nil
}

func getGatewayProtocolType(protocol gatewayv1.ProtocolType) (string, error) {
	switch protocol {
	case gatewayv1.HTTPProtocolType:
		return "http", nil
	case gatewayv1.HTTPSProtocolType:
		return "https", nil
	case gatewayv1.TCPProtocolType:
		return "tcp", nil
	case gatewayv1.TLSProtocolType:
		return "https", nil
	case gatewayv1.UDPProtocolType:
		return "udp", nil
	default:
		return "", fmt.Errorf("Unknown Protocol type: %s", protocol)
	}
}

type HttpRouteMap map[string]*AnnotatedHTTPRoute

// adds route to HttpRouteMap
// if route exists, it adds the gateway to the annotated routes gateways
func (route_map HttpRouteMap) AddUnique(route *gatewayv1.HTTPRoute, gateways []*gatewayv1.Gateway) {
	if _, ok := route_map[route.Namespace+"/"+route.Name]; !ok {
		route_map[route.Namespace+"/"+route.Name] = &AnnotatedHTTPRoute{
			HTTPRoute: route,
			gateways:  GatewayMap{},
		}
	}

	if len(gateways) > 0 {
		for _, gateway := range gateways {
			route_map[route.Namespace+"/"+route.Name].addGateway(gateway)
		}
	}
}

// NOTE: single value fields (bools, strings) are set by first non nil (order is route -> gateway in order of parentRefs)
// NOTE: merge behavior of multi value fields (maps, lists) can be configured (TODO: make this actually configurable)
// TODO: add ability to toggle wether to create one endpoint per detected hostname or pick first
func (obj *AnnotatedHTTPRoute) GetEndpointConfigs(config config.Config) ([]*gatusconfig.GatusEndpointConfig, error) {
	route_hostnames := obj.getHostnames()

	hostname_config_map := map[string][]*gatusconfig.GatusEndpointConfig{}

	// get different URLs and preferred protocol + merged config

	route_config, err := parseGatusAnnotations(obj.HTTPRoute)

	if err != nil {
		return nil, fmt.Errorf("Unable to parse HTTPRoute Gatus annotations: %s", err)
	}

	protocols := map[string]struct{}{}

	for _, parent_ref := range obj.Spec.ParentRefs {
		gateway, listener_name := obj.getGateway(parent_ref)
		if gateway == nil {
			continue
		}
		for _, listener := range gateway.Spec.Listeners {
			listener_hostnames := route_hostnames

			if listener.Hostname != nil && (len(route_hostnames) == 0 || slices.Contains(listener_hostnames, string(*listener.Hostname))) {
				listener_hostnames = []string{string(*listener.Hostname)}
			}

			protocol, err := getGatewayProtocolType(listener.Protocol)
			if err != nil {
				return nil, fmt.Errorf("Could not get correct protocol type for route: %s", err)
			}
			protocols[protocol] = struct{}{}

			if listener_name == nil || listener.Name == *listener_name {
				for _, hostname := range listener_hostnames {
					if _, ok := hostname_config_map[hostname]; !ok {
						hostname_config_map[hostname] = []*gatusconfig.GatusEndpointConfig{}
					}
					parsed_config, err := parseGatusAnnotations(gateway)
					if err != nil {
						return nil, fmt.Errorf("Unable to parse Gateway Gatus annotations: %s", err)
					}
					if parsed_config != nil {
						parsed_config.URL = hostname
						hostname_config_map[hostname] = append(hostname_config_map[hostname], parsed_config)
					}
				}
			}
		}
	}

	protocol, err := getPreferredProtocol(config, protocols)
	if err != nil {
		return nil, fmt.Errorf("Could not get preferred Protocol for HTTPRoute: %s", err)
	}

	paths := obj.getPaths()
	configs := make([]*gatusconfig.GatusEndpointConfig, 0)
	for hostname, cfgs := range hostname_config_map {
		merged_cfg := route_config.Merge(cfgs...)

		for path := range paths {
			cfg := merged_cfg.Clone()

			cfg.URL = protocol + "://" + cfg.URL + path

			err := defaultConfig(cfg, obj)

			// make name unique if multiple Hostnames exist for the same httproute and therefore name
			if len(hostname_config_map) > 1 {
				merged_cfg.Name = merged_cfg.Name + " - " + hostname
			}
			// make name unique if multiple Protocols exist for the same httproute and therefore name
			if len(paths) > 1 {
				cfg.Name = cfg.Name + " - " + path
			}

			if err != nil {
				return nil, err
			}

			configs = append(configs, cfg)
		}
	}
	return configs, nil
}
