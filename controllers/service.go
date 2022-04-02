package controllers

import (
	appv1 "nginx-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

// serviceForOperator returns an nginxoperator Service object
func (r *NginxOperatorReconciler) serviceForOperator(m *appv1.NginxOperator) *corev1.Service {
	ls := labelsForOperator(m.Name)

	dep := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{
				{
					Port: 5678,
					TargetPort: intstr.IntOrString{
						IntVal: 5678,
					},
					Protocol: "TCP",
				},
			},
		},
	}
	// Set Operator instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}
