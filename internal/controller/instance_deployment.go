package controller

import (
	"fmt"
	"maps"

	gatusiov1alpha1 "github.com/Jannik-Hm/Gatus-Operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getInstanceLabels(instance *gatusiov1alpha1.Instance) map[string]string {
	return map[string]string{
		"app.kubernetes.io/instance": instance.Name,
		"managed-by":                 "gatus-operator",
	}
}

func getDeploymentLabels(instance *gatusiov1alpha1.Instance) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "gatus",
		"app.kubernetes.io/instance": instance.Name,
		"managed-by":                 "gatus-operator",
	}
}

func mutateDeployment(instance *gatusiov1alpha1.Instance, obj *appsv1.Deployment, scheme *runtime.Scheme, configmap *corev1.ConfigMap) error {
	// Check if the object already exists and has a different owner
	if obj.GetResourceVersion() != "" { // Object exists
		owner := metav1.GetControllerOf(obj)
		if owner == nil || owner.UID != instance.UID {
			return fmt.Errorf("deployment %s already exists and is not managed by this operator", obj.GetName())
		}
	}

	// Set/Ensure the Controller Reference
	if err := controllerutil.SetControllerReference(instance, obj, scheme); err != nil {
		return err
	}

	var deploymentLabels = getDeploymentLabels(instance)

	if instance.Spec.Image == nil {
		return fmt.Errorf("Image is empty, this means that the defaulting webhook is not working properly.")
	}
	image := *instance.Spec.Image

	obj.Spec.Replicas = instance.Spec.Replicas

	obj.Spec.ProgressDeadlineSeconds = ptr.To[int32](120) // rollout should fail after 120 seconds (gatus has quick startup and fails fast if cfg is wrong)
	// TODO: either fine tune this setting or make it available to configure via env var

	obj.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: maps.Clone(deploymentLabels),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "gatus",
					Image: image,
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 8080,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					SecurityContext: &corev1.SecurityContext{
						ReadOnlyRootFilesystem: ptr.To(true),
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{
							Path: "/health",
							Port: intstr.FromString("http"),
						}},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{
							Path: "/health",
							Port: intstr.FromString("http"),
						}},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "gatus-config",
							MountPath: "/config",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "gatus-config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: configmap.Name,
							},
						},
					},
				},
			},
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot: ptr.To(true),
				RunAsUser:    ptr.To[int64](65534),
				RunAsGroup:   ptr.To[int64](65534),
				FSGroup:      ptr.To[int64](65534),
			},
		},
	}

	// We only set the selector IF the object is being created.
	// If it exists, we leave the selector alone to avoid immutability errors.
	if obj.ObjectMeta.CreationTimestamp.IsZero() {
		obj.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: maps.Clone(deploymentLabels),
		}
	}

	return nil
}

func mutateService(instance *gatusiov1alpha1.Instance, obj *corev1.Service, scheme *runtime.Scheme) error {
	// Check if the object already exists and has a different owner
	if obj.GetResourceVersion() != "" { // Object exists
		owner := metav1.GetControllerOf(obj)
		if owner == nil || owner.UID != instance.UID {
			return fmt.Errorf("deployment %s already exists and is not managed by this operator", obj.GetName())
		}
	}

	// Set/Ensure the Controller Reference
	if err := controllerutil.SetControllerReference(instance, obj, scheme); err != nil {
		return err
	}

	var deploymentLabels = getDeploymentLabels(instance)
	var instanceLabels = getInstanceLabels(instance)

	var serviceLabels = map[string]string{}
	maps.Copy(serviceLabels, instance.Spec.Service.ServiceLabels)
	maps.Copy(serviceLabels, instanceLabels)

	obj.Annotations = maps.Clone(instance.Spec.Service.ServiceAnnotations)
	obj.Labels = maps.Clone(serviceLabels)

	obj.Spec = corev1.ServiceSpec{
		Selector: maps.Clone(deploymentLabels),
		Ports: []corev1.ServicePort{
			{
				Name:       "http",
				Protocol:   corev1.ProtocolTCP,
				Port:       80,
				TargetPort: intstr.FromString("http"),
			},
		},
		Type:           instance.Spec.Service.ServiceType,
		IPFamilyPolicy: instance.Spec.Service.IPFamilyPolicy,
		IPFamilies:     instance.Spec.Service.IPFamilies,
	}

	return nil
}
