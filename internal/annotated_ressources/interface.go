package annotatedressources

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AnnotatedRessource interface {
	metav1.Object
	// TODO: add unique key that gets appended to endpoint name if more than one entry
	GetURLs() ([]string, error)
	GetConditions(protocol string) []string
}
