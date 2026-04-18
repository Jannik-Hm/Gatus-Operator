package annotatedressources

import (
	"fmt"
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
	annotationPrefix string = ".metadata.annotations."

	disabledAnnotation       string = "gatus.io/disabled"
	nameAnnotation           string = "gatus.io/name"
	instancesAnnotation      string = "gatus.io/instances"
	groupAnnotation          string = "gatus.io/group"
	configOverrideAnnotation string = "gatus.io/config"

	disabledAnnotationWithPrefix       string = annotationPrefix + disabledAnnotation
	nameAnnotationWithPrefix           string = annotationPrefix + nameAnnotation
	instancesAnnotationWithPrefix      string = annotationPrefix + instancesAnnotation
	groupAnnotationWithPrefix          string = annotationPrefix + groupAnnotation
	configOverrideAnnotationWithPrefix string = annotationPrefix + configOverrideAnnotation

	gatewayParentRefSpec string = "gatewayParentRef"
)

func parseGatusAnnotations(obj client.Object) (*gatusconfig.GatusEndpointConfig, error) {
	annotations := obj.GetAnnotations()
	var cfg gatusconfig.GatusEndpointConfig
	if additionalConfigString, ok := annotations[configOverrideAnnotation]; ok {
		err := yaml.Unmarshal([]byte(additionalConfigString), &cfg)
		if err != nil {
			return nil, fmt.Errorf("Could not parse override config annotation: %s", err)
		}
	} else {
		cfg = gatusconfig.GatusEndpointConfig{}
	}

	if name, ok := annotations[nameAnnotation]; ok && name != "" {
		cfg.Name = name
	}
	if group, ok := annotations[groupAnnotation]; ok {
		cfg.Group = &group
	}

	// TODO: add the rest of the annotations

	return &cfg, nil
}

func annotationsToGatusConfigs(obj AnnotatedRessource) ([]*gatusconfig.GatusEndpointConfig, error) {
	annotations := obj.GetAnnotations()
	var base_cfg gatusconfig.GatusEndpointConfig
	if additionalConfigString, ok := annotations[configOverrideAnnotation]; ok {
		err := yaml.Unmarshal([]byte(additionalConfigString), &base_cfg)
		if err != nil {
			return nil, fmt.Errorf("Could not parse override config annotation: %s", err)
		}
	} else {
		base_cfg = gatusconfig.GatusEndpointConfig{}
	}

	urls, err := obj.GetURLs()
	if err != nil {
		return nil, fmt.Errorf("Could not generate URLs: %s", err)
	}
	configs := make([]*gatusconfig.GatusEndpointConfig, 0, len(urls))

	for _, url := range urls {
		cfg := base_cfg

		cfg.URL = url

		if len(cfg.Conditions) == 0 {
			protocol, _, _ := strings.Cut(url, "://")
			cfg.Conditions = obj.GetConditions(protocol)
		}

		if name, ok := annotations[nameAnnotation]; ok && name != "" {
			cfg.Name = name
		} else {
			cfg.Name = obj.GetName()
		}
		// TODO: append unique string to name if len(urls) > 1

		if group, ok := annotations[groupAnnotation]; ok {
			cfg.Group = &group
		} else {
			cfg.Group = ptr.To(obj.GetNamespace())
		}
		configs = append(configs, &cfg)
	}

	return configs, nil
}
