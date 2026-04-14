package annotatedressources

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AnnotatedRessource interface {
	metav1.Object
	GetURLs() []string
	GetConditions() []string
}
