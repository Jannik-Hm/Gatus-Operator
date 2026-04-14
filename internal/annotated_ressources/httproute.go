package annotatedressources

import (
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type AnnotatedHTTPRoute gatewayv1.HTTPRoute

func (obj *AnnotatedHTTPRoute) GetURLs() []string {
	urls := make([]string, 0)

	// TODO: make this work
	// paths := obj.getPaths()

	// for _, hostname := range obj.getHostnames() {
	// }

	// TODO: to check wether endpoint is HTTP or HTTPS, Gateway Spec needs to be checked

	return urls
}

func (obj *AnnotatedHTTPRoute) GetConditions() []string {
	return []string{"[STATUS] == 200"}
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
