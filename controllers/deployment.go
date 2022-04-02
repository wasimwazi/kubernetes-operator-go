package controllers

import (
	appv1 "nginx-operator/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// deploymentForOperator returns an nginxoperator Deployment object
func (r *NginxOperatorReconciler) deploymentForOperator(m *appv1.NginxOperator) *appsv1.Deployment {
	ls := labelsForOperator(m.Name)
	replicas := m.Spec.Replicas

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image:           m.Spec.Image,
							Args:            []string{"-text=foo"},
							ImagePullPolicy: corev1.PullAlways,
							Name:            "operator-image",
							Ports: []corev1.ContainerPort{{
								ContainerPort: 5678,
							}},
						},
					},
				},
			},
		},
	}

	// Set NginxOperator instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}
