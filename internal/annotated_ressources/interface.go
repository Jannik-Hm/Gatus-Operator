package annotatedressources

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Jannik-Hm/Gatus-Operator/internal/config"
	gatusconfig "github.com/Jannik-Hm/Gatus-Operator/internal/gatus_config"
	"go.yaml.in/yaml/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AnnotatedRessource interface {
	metav1.Object
	// TODO: add unique key that gets appended to endpoint name if more than one entry
	GetURLs() ([]string, error)
	GetConditions(protocol string) []string
	GetEndpointConfigs(config config.Config) ([]*gatusconfig.GatusEndpointConfig, error)
}

func defaultConfig(cfg *gatusconfig.GatusEndpointConfig, obj AnnotatedRessource) error {
	if len(cfg.Conditions) == 0 {
		protocol, _, _ := strings.Cut(cfg.URL, "://")
		cfg.Conditions = obj.GetConditions(protocol)
		if len(cfg.Conditions) == 0 {
			return fmt.Errorf("Generated 0 Conditions for Endpoint %s; This is the protocol: %s", cfg.Name, protocol)
		}
	}

	if cfg.Name == "" {
		cfg.Name = obj.GetName()
	}
	if cfg.Group == nil {
		cfg.Group = ptr.To(obj.GetNamespace())
	}

	return nil
}

const (
	DisabledAnnotation       string = "gatus.io/disabled"
	NameAnnotation           string = "gatus.io/name"
	InstancesAnnotation      string = "gatus.io/instances"
	GroupAnnotation          string = "gatus.io/group"
	ConfigOverrideAnnotation string = "gatus.io/config"
)

func parseGatusAnnotations(obj client.Object) (*gatusconfig.GatusEndpointConfig, error) {
	// TODO: add option to disallow annotations: e.g. a name annotation on a gateway does not make sense (uniqueness)
	annotations := obj.GetAnnotations()
	var cfg gatusconfig.GatusEndpointConfig
	if additionalConfigString, ok := annotations[ConfigOverrideAnnotation]; ok {
		err := yaml.Unmarshal([]byte(additionalConfigString), &cfg)
		if err != nil {
			return nil, fmt.Errorf("Could not parse override config annotation: %s", err)
		}
	} else {
		cfg = gatusconfig.GatusEndpointConfig{}
	}

	if name, ok := annotations[NameAnnotation]; ok && name != "" {
		cfg.Name = name
	}
	if group, ok := annotations[GroupAnnotation]; ok {
		cfg.Group = &group
	}

	// TODO: add the rest of the annotations

	return &cfg, nil
}

func getPreferredProtocol(config config.Config, protocols map[string]struct{}) (string, error) {
	var protocol string
	for _, preferred_protocol := range config.ProtocolPreference {
		if _, ok := protocols[preferred_protocol]; ok {
			protocol = preferred_protocol
		}
	}
	if protocol == "" {
		return "", fmt.Errorf("Preferred Protocols list does not contain detected protocols")
	}
	return protocol, nil
}

var hostname_path_regex = regexp.MustCompile(`^(?:[a-zA-Z0-9+-.]+:\/\/)?(?:[^@\n]+@)?(.*[\n?]+)`)

func getHostnameAndPathFromUrl(url string) (string, error) {
	matches := hostname_path_regex.FindStringSubmatch(url)
	if len(matches) == 1 {
		return matches[0], nil
	}
	return "", fmt.Errorf("Regex could not find a hostname in the specified url: %s", url)
}
