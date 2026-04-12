package v1alpha1

// +kubebuilder:object:generate=false
type ConfigConverter[T any] interface {
	ToGatusConfig() T
}

func ToGatusConfigList[Source ConfigConverter[Destination], Destination any](source []Source) []Destination {
	result := make([]Destination, 0, len(source))
	for _, value := range source {
		result = append(result, value.ToGatusConfig())
	}
	return result
}

func ToGatusConfigMap[Source ConfigConverter[Destination], Destination any](source map[string]Source) map[string]Destination {
	result := map[string]Destination{}
	for key, value := range source {
		result[key] = value.ToGatusConfig()
	}
	return result
}
