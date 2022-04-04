package controllers

import (
	appv1 "nginx-operator/api/v1"

	v1 "github.com/jetstack/cert-manager/pkg/apis/acme/v1"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// issuerForIngress returns a issuer object
func (r *NginxOperatorReconciler) issuerForIngress(m *appv1.NginxOperator, issuerName string) *cmapi.Issuer {
	serverURL := "https://acme-v02.api.letsencrypt.org/directory"
	email := "wasimakr@cisco.com"
	ingressClass := "nginx"
	issuer := cmapi.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      issuerName,
			Namespace: m.Namespace,
		},
		Spec: cmapi.IssuerSpec{
			IssuerConfig: cmapi.IssuerConfig{
				ACME: &v1.ACMEIssuer{
					Server: serverURL,
					Email:  email,
					PrivateKey: cmmetav1.SecretKeySelector{
						LocalObjectReference: cmmetav1.LocalObjectReference{
							Name: "operator-secret-key",
						},
					},
					Solvers: []v1.ACMEChallengeSolver{
						{
							HTTP01: &v1.ACMEChallengeSolverHTTP01{
								Ingress: &v1.ACMEChallengeSolverHTTP01Ingress{
									Class: &ingressClass,
								},
							},
						},
					},
				},
			},
		},
	}

	// Set Operator instance as the owner and controller
	ctrl.SetControllerReference(m, &issuer, r.Scheme)
	return &issuer
}

// certificateForIngress returns a certificate object
func (r *NginxOperatorReconciler) certificateForIngress(m *appv1.NginxOperator, issuerName, certName string) *cmapi.Certificate {
	secretName := m.Name + "-secret"
	certificate := cmapi.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certName,
			Namespace: m.Namespace,
		},
		Spec: cmapi.CertificateSpec{
			SecretName: secretName,
			IssuerRef: cmmetav1.ObjectReference{
				Name: issuerName,
			},
			DNSNames: []string{m.Spec.Host},
		},
	}

	// Set Operator instance as the owner and controller
	ctrl.SetControllerReference(m, &certificate, r.Scheme)
	return &certificate
}

// ingressForOperator returns a nginxoperator Ingress object
func (r *NginxOperatorReconciler) ingressForOperator(m *appv1.NginxOperator, issuerName, certName string) *networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix
	ingressPaths := []networkingv1.HTTPIngressPath{
		{
			Path:     "/",
			PathType: &pathType,
			Backend: networkingv1.IngressBackend{
				Service: &networkingv1.IngressServiceBackend{
					Name: m.Name,
					Port: networkingv1.ServiceBackendPort{
						Number: 5678,
					},
				},
			},
		},
	}
	ingressSpec := networkingv1.IngressSpec{
		TLS: []networkingv1.IngressTLS{
			{
				Hosts:      []string{m.Spec.Host},
				SecretName: certName,
			},
		},
		Rules: []networkingv1.IngressRule{
			{
				Host: m.Spec.Host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: ingressPaths,
					},
				},
			},
		},
	}
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":                    "nginx",
				"nginx.ingress.kubernetes.io/ssl-redirect":       "true",
				"nginx.ingress.kubernetes.io/force-ssl-redirect": "true",
				"cert-manager.io/issuer":                         issuerName,
			},
		},
		Spec: ingressSpec,
	}

	// Set Operator instance as the owner and controller
	ctrl.SetControllerReference(m, ingress, r.Scheme)
	return ingress
}
