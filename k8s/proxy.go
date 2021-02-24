package k8s

import (
	managedv1alpha1 "github.com/dedgar/traffic-gen-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ProxyDeployment returns a new daemonset customized for Proxy
func ProxyDeployment(m *managedv1alpha1.Proxy) *appsv1.Deployment {
	var privileged = true
	var runAsUser int64

	ds := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "Proxy",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "Proxy",
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"node-role.kubernetes.io/master": "",
					},
					// ServiceAccountName: "openshift-traffic-gen-operator",
					Tolerations: []corev1.Toleration{
						{
							Operator: corev1.TolerationOpExists,
						},
					},
					Containers: []corev1.Container{{
						Image: "quay.io/dedgar/pod-Proxy:v0.0.10",
						Name:  "Proxy",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privileged,
							RunAsUser:  &runAsUser,
						},
						Env: []corev1.EnvVar{{
							Name:  "OO_PAUSE_ON_START",
							Value: "false",
						}, {
							Name:  "LOG_WRITER_URL",
							Value: "http://Proxy.openshift-traffic-gen-operator.svc:8080/api/log",
						}, {
							Name:  "SCAN_LOG_FILE",
							Value: "/host/var/log/openshift_managed_malware_scan.log",
						}, {
							Name:  "POD_LOG_FILE",
							Value: "/host/var/log/openshift_managed_pod_creation.log",
						}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "Proxy",
						}},
						Resources: corev1.ResourceRequirements{},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "Proxy-secrets",
							MountPath: "/secrets",
						}, {
							Name:      "host-logs",
							MountPath: "/host/var/log/",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "Proxy-secrets",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "Proxy-secrets",
							},
						},
					}, {
						Name: "host-logs",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/var/log/",
							},
						},
					}},
				},
			},
		},
	}
	return ds
}

// ProxyService returns a new service customized for Proxy
func ProxyService(m *managedv1alpha1.ProxyService) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels: map[string]string{
				"name":    m.Name,
				"k8s-app": m.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"name": "Proxy",
			},
			Ports: []corev1.ServicePort{{
				Port:       8080,
				TargetPort: intstr.FromInt(8080),
				Name:       m.Name,
			}},
		},
	}
	return svc
}
