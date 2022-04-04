package controllers

import (
	"context"

	"github.com/go-logr/logr"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1 "nginx-operator/api/v1"
)

// NginxOperatorReconciler reconciles a NginxOperator object
type NginxOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.cisco.com,resources=nginxoperators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.cisco.com,resources=nginxoperators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.cisco.com,resources=nginxoperators/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NginxOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.Log.WithName("NginxOperator")
	// Fetch the Operator instance
	operator := &appv1.NginxOperator{}
	err := r.Get(ctx, req.NamespacedName, operator)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("NginxOperator resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Operator")
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	foundDeployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: operator.Name, Namespace: operator.Namespace}, foundDeployment)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForOperator(operator)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		log.Info("Deployment created successfully", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Check if the deployment Spec.Template, matches the found Spec.Template
	deploy := r.deploymentForOperator(operator)
	if !equality.Semantic.DeepDerivative(deploy.Spec.Template, foundDeployment.Spec.Template) {
		foundDeployment = deploy
		log.Info("Updating Deployment", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
		err := r.Update(ctx, foundDeployment)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
			return ctrl.Result{}, err
		}
		log.Info("Successfully Updated Deployment", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
		return ctrl.Result{Requeue: true}, nil
	}

	// Ensure the deployment size is the same as the spec
	size := operator.Spec.Replicas
	if *foundDeployment.Spec.Replicas != size {
		foundDeployment.Spec.Replicas = &size
		log.Info("Updating Deployment for matching replicas", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
		err = r.Update(ctx, foundDeployment)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
			return ctrl.Result{}, err
		}
		log.Info("Successfully Updated Deployment", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Check if the service already exists, if not create a new one
	foundService := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: operator.Name, Namespace: operator.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		svc := r.serviceForOperator(operator)
		log.Info("Creating a new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return ctrl.Result{}, err
		}
		log.Info("Service creation successfull", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	// Create issuer for ingress
	issuerName := operator.Name + "-issuer"
	foundIssuer := &cmapi.Issuer{}
	err = r.Get(ctx, types.NamespacedName{Namespace: operator.Namespace, Name: issuerName}, foundIssuer)
	if err != nil && errors.IsNotFound(err) {
		// Define a new issuer
		iss := r.issuerForIngress(operator, issuerName)
		log.Info("Creating a new issuer", "Issuer.Namespace", iss.Namespace, "Issuer.Name", iss.Name)
		err = r.Create(ctx, iss)
		if err != nil {
			log.Error(err, "Failed to create new Issuer", "Issuer.Namespace", iss.Namespace, "Issuer.Name", iss.Name)
			return ctrl.Result{}, err
		}
		log.Info("Issuer created successfully", "Issuer.Namespace", iss.Namespace, "Issuer.Name", iss.Name)
		// Issuer created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Issuer")
		return ctrl.Result{}, err
	}

	// Create certificate for ingress
	certName := operator.Name + "-certificate"
	foundCertificate := &cmapi.Certificate{}
	err = r.Get(ctx, types.NamespacedName{Namespace: operator.Namespace, Name: certName}, foundCertificate)
	if err != nil && errors.IsNotFound(err) {
		// Define a new certificate
		cer := r.certificateForIngress(operator, foundIssuer.Name, certName)
		log.Info("Creating a new certificate", "Certificate.Namespace", cer.Namespace, "Certificate.Name", cer.Name)
		err = r.Create(ctx, cer)
		if err != nil {
			log.Error(err, "Failed to create new Certificate", "Certificate.Namespace", cer.Namespace, "Certificate.Name", cer.Name)
			return ctrl.Result{}, err
		}
		log.Info("Certificate created successfully", "Certificate.Namespace", cer.Namespace, "Certificate.Name", cer.Name)
		// Certificate created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Certificate")
		return ctrl.Result{}, err
	}

	// Check if the ingress already exists, if not create a new one
	foundIngress := &networkingv1.Ingress{}
	err = r.Get(ctx, types.NamespacedName{Name: operator.Name, Namespace: operator.Namespace}, foundIngress)
	if err != nil && errors.IsNotFound(err) {
		// Define a new ingress
		ing := r.ingressForOperator(operator, foundIssuer.Name, foundCertificate.Name)
		log.Info("Creating a new Ingress", "Ingress.Namespace", ing.Namespace, "Ingress.Name", ing.Name)
		err = r.Create(ctx, ing)
		if err != nil {
			log.Error(err, "Failed to create new Ingress", "Ingress.Namespace", ing.Namespace, "Ingress.Name", ing.Name)
			return ctrl.Result{}, err
		}
		log.Info("Ingress created successfully", "Ingress.Namespace", ing.Namespace, "Ingress.Name", ing.Name)
		// Ingress created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Ingress")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// labelsForOperator returns the labels for selecting the resources
// belonging to the given nginxoperator CR name.
func labelsForOperator(name string) map[string]string {
	return map[string]string{"app": "operator", "cr_name": name}
}

func (r *NginxOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.NginxOperator{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}
