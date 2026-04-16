package annotatedressources

import (
	"context"
	"fmt"
	"slices"

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

func (obj *AnnotatedHTTPRoute) GetEndpointConfigs() ([]*gatusconfig.GatusEndpointConfig, error) {
	urls := obj.getHostnames()

	url_config_map := map[string][]*gatusconfig.GatusEndpointConfig{}

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
					// TODO: this url is not assembled fully yet (missing protocol and paths)
					if _, ok := url_config_map[url]; !ok {
						url_config_map[url] = make([]*gatusconfig.GatusEndpointConfig, 0)
					}
					parsed_config, err := parseGatusAnnotations(gateway)
					if err != nil {
						return nil, fmt.Errorf("Unable to parse Gatus annotations: %s", err)
					}
					url_config_map[url] = append(url_config_map[url], parsed_config)
				}
			}
		}
	}
	configs := make([]*gatusconfig.GatusEndpointConfig, 0)
	for url, cfg := range url_config_map {
		// TODO: merge cfgs and append to `configs`
		_, _ = url, cfg
	}
	return configs, nil
}
