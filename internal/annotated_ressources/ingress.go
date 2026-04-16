package annotatedressources

import (
	"fmt"

	gatusconfig "github.com/Jannik-Hm/Gatus-Operator/internal/gatus_config"
	networkingv1 "k8s.io/api/networking/v1"
)

type AnnotatedIngress networkingv1.Ingress

func (obj *AnnotatedIngress) GetURLs() ([]string, error) {
	urls := make([]string, 0)

	tls_enabled_hosts := obj.getTLSHosts()
	for host, host_paths := range obj.getHostnamesWithPath() {
		for _, host_path := range host_paths {
			protocol := "http"
			if _, ok := tls_enabled_hosts[host]; ok {
				protocol = "https"
			}
			urls = append(urls, fmt.Sprintf("%s://%s%s", protocol, host, host_path))
		}
	}

	return urls, nil
}

func (obj *AnnotatedIngress) GetConditions(protocol string) []string {
	return []string{"[STATUS] == 200"}
}

func (obj *AnnotatedIngress) getHostnamesWithPath() map[string][]string {
	hostnamesWithPath := make(map[string][]string)
	for _, rule := range obj.Spec.Rules {
		if rule.Host != "" {
			for _, path := range rule.HTTP.Paths {
				if _, ok := hostnamesWithPath[rule.Host]; !ok {
					hostnamesWithPath[rule.Host] = make([]string, 0)
				}
				hostnamesWithPath[rule.Host] = append(hostnamesWithPath[rule.Host], path.Path)
			}
		}
	}
	return hostnamesWithPath
}

func (obj *AnnotatedIngress) getTLSHosts() map[string]struct{} {
	tls_enabled_hosts := map[string]struct{}{}
	for _, cert := range obj.Spec.TLS {
		for _, host := range cert.Hosts {
			tls_enabled_hosts[host] = struct{}{}
		}
	}
	return tls_enabled_hosts
}

func (obj *AnnotatedIngress) GetEndpointConfigs() ([]*gatusconfig.GatusEndpointConfig, error)
