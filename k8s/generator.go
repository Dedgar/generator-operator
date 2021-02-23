package k8s

import (
	managedv1alpha1 "github.com/dedgar/traffic-generator-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GeneratorDaemonSet returns a new daemonset customized for generator
func GeneratorDaemonSet(m *managedv1alpha1.Generator) *appsv1.DaemonSet {
	var privileged = true
	var runAsUser int64

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "generator",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "generator",
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"node-role.kubernetes.io/master": "",
					},
					// ServiceAccountName: "openshift-traffic-generator-operator",
					Tolerations: []corev1.Toleration{
						{
							Operator: corev1.TolerationOpExists,
						},
					},
					Containers: []corev1.Container{{
						Image: "quay.io/dedgar/generator:v0.0.1",
						Name:  "generator",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privileged,
							RunAsUser:  &runAsUser,
						},
						Env: []corev1.EnvVar{{
							Name:  "OO_PAUSE_ON_START",
							Value: "false",
						}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "generator",
						}},
						Resources: corev1.ResourceRequirements{},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "generator-secrets",
							MountPath: "/secrets",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "generator-secrets",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "generator-secrets",
							},
						},
					}},
				},
			},
		},
	}
	return ds
}
