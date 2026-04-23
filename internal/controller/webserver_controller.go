package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	webv1 "github.com/XiongGG1989/web-operator/api/v1"
)

// WebServerReconciler reconciles a WebServer object.
type WebServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    klog.Logger
}

// +kubebuilder:rbac:groups=web.xm.web,resources=webservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=web.xm.web,resources=webservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *WebServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("webserver", req.NamespacedName)

	ws := &webv1.WebServer{}
	if err := r.Get(ctx, req.NamespacedName, ws); err != nil {
		if errors.IsNotFound(err) {
			log.Info("WebServer resource deleted", "namespace", req.Namespace, "name", req.Name)
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get WebServer")
		return ctrl.Result{}, err
	}

	cfg := buildWebServerConfig(ws)

	deployResult, deployChanged, err := r.reconcileDeployment(ctx, ws, cfg)
	if err != nil {
		if errors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		log.Error(err, "Failed to reconcile Deployment")
		return ctrl.Result{}, err
	}
	r.logDeploymentResult(log, ws, cfg, deployResult, deployChanged)

	serviceResult, serviceChanged, err := r.reconcileService(ctx, ws, cfg)
	if err != nil {
		if errors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		log.Error(err, "Failed to reconcile Service")
		return ctrl.Result{}, err
	}
	r.logServiceResult(log, ws, cfg, serviceResult, serviceChanged)

	if err := r.reconcileStatus(ctx, log, ws, cfg); err != nil {
		if errors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		log.Error(err, "Failed to update WebServer status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *WebServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log = klog.LoggerWithName(klog.Background(), "webserver-controller")
	return ctrl.NewControllerManagedBy(mgr).
		For(&webv1.WebServer{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
