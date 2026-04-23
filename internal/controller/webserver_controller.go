package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	webv1 "github.com/XiongGG1989/web-operator/api/v1"
)

// WebServerReconciler reconciles a WebServer object
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

	// 确定目标 namespace
	targetNS := ws.Namespace
	if ws.Spec.TargetNamespace != "" {
		targetNS = ws.Spec.TargetNamespace
	}

	// 默认值
	replicas := int32(1)
	if ws.Spec.Replicas != nil {
		replicas = *ws.Spec.Replicas
	}
	image := ws.Spec.Image
	if image == "" {
		image = "nginx:latest"
	}
	port := ws.Spec.Port
	if port == 0 {
		port = 80
	}

	svcType := corev1.ServiceTypeClusterIP
	switch ws.Spec.ServiceType {
	case "NodePort":
		svcType = corev1.ServiceTypeNodePort
	case "LoadBalancer":
		svcType = corev1.ServiceTypeLoadBalancer
	}

	labels := map[string]string{
		"app.kubernetes.io/name":     "webserver",
		"app.kubernetes.io/instance": ws.Name,
	}

	// Reconcile Deployment
	deploy := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: ws.Name, Namespace: targetNS}}
	var deployChanged bool
	deployResult, err := controllerutil.CreateOrUpdate(ctx, r.Client, deploy, func() error {
		// 检测是否有实际变化
		if deploy.ResourceVersion != "" {
			if deploy.Spec.Replicas != nil && *deploy.Spec.Replicas != replicas {
				deployChanged = true
			}
			if len(deploy.Spec.Template.Spec.Containers) > 0 {
				if deploy.Spec.Template.Spec.Containers[0].Image != image {
					deployChanged = true
				}
				if len(deploy.Spec.Template.Spec.Containers[0].Ports) > 0 &&
					deploy.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort != port {
					deployChanged = true
				}
			}
		}

		deploy.Labels = labels
		deploy.Spec = appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:      "web",
						Image:     image,
						Resources: ws.Spec.Resources,
						Ports:     []corev1.ContainerPort{{ContainerPort: port}},
					}},
				},
			},
		}
		return ctrl.SetControllerReference(ws, deploy, r.Scheme)
	})
	if err != nil {
		if errors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		log.Error(err, "Failed to reconcile deployment")
		return ctrl.Result{}, err
	}
	if deployResult == controllerutil.OperationResultCreated {
		log.Info("Deployment created", "namespace", targetNS, "name", ws.Name, "replicas", replicas, "image", image)
	} else if deployResult == controllerutil.OperationResultUpdated && deployChanged {
		log.Info("Deployment updated", "namespace", targetNS, "name", ws.Name)
	}

	// Reconcile Service
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: ws.Name, Namespace: targetNS}}
	var svcChanged bool
	svcResult, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		// 检测是否有实际变化
		if svc.ResourceVersion != "" {
			if len(svc.Spec.Ports) > 0 && svc.Spec.Ports[0].Port != port {
				svcChanged = true
			}
			if svc.Spec.Type != svcType {
				svcChanged = true
			}
		}

		svc.Labels = labels
		svc.Spec = corev1.ServiceSpec{
			Type:     svcType,
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Port:       port,
				TargetPort: intstr.FromInt(int(port)),
				Protocol:   corev1.ProtocolTCP,
			}},
		}
		return ctrl.SetControllerReference(ws, svc, r.Scheme)
	})
	if err != nil {
		if errors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		log.Error(err, "Failed to reconcile service")
		return ctrl.Result{}, err
	}
	if svcResult == controllerutil.OperationResultCreated {
		log.Info("Service created", "namespace", targetNS, "name", ws.Name, "type", svcType, "port", port)
	} else if svcResult == controllerutil.OperationResultUpdated && svcChanged {
		log.Info("Service updated", "namespace", targetNS, "name", ws.Name)
	}

	// Update Status
	currentDeploy := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: ws.Name, Namespace: targetNS}, currentDeploy); err != nil {
		setCondition(&ws.Status, metav1.Condition{
			Type:    "Degraded",
			Status:  metav1.ConditionTrue,
			Reason:  "DeploymentNotFound",
			Message: err.Error(),
		})
	} else {
		// 记录 Pod 数量变化
		if currentDeploy.Status.AvailableReplicas != ws.Status.AvailableReplicas {
			log.Info("Pod replicas changed",
				"namespace", targetNS,
				"name", ws.Name,
				"available", currentDeploy.Status.AvailableReplicas,
				"desired", *currentDeploy.Spec.Replicas,
				"ready", currentDeploy.Status.ReadyReplicas,
				"updated", currentDeploy.Status.UpdatedReplicas)
		}

		ws.Status.AvailableReplicas = currentDeploy.Status.AvailableReplicas

		setCondition(&ws.Status, metav1.Condition{
			Type:    "Available",
			Status:  boolToCondition(currentDeploy.Status.AvailableReplicas > 0),
			Reason:  map[bool]string{true: "DeploymentReady", false: "NoAvailableReplicas"}[currentDeploy.Status.AvailableReplicas > 0],
			Message: map[bool]string{true: "Deployment has available replicas", false: "No available replicas"}[currentDeploy.Status.AvailableReplicas > 0],
		})

		updating := currentDeploy.Spec.Replicas != nil && currentDeploy.Status.UpdatedReplicas != *currentDeploy.Spec.Replicas
		setCondition(&ws.Status, metav1.Condition{
			Type:    "Progressing",
			Status:  boolToCondition(updating),
			Reason:  map[bool]string{true: "Updating", false: "UpToDate"}[updating],
			Message: map[bool]string{true: "Deployment is updating", false: "Deployment is up-to-date"}[updating],
		})
	}

	if err := r.Status().Update(ctx, ws); err != nil {
		if errors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func setCondition(status *webv1.WebServerStatus, condition metav1.Condition) {
	meta.SetStatusCondition(&status.Conditions, condition)
}

func boolToCondition(b bool) metav1.ConditionStatus {
	if b {
		return metav1.ConditionTrue
	}
	return metav1.ConditionFalse
}

func (r *WebServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log = klog.LoggerWithName(klog.Background(), "webserver-controller")
	return ctrl.NewControllerManagedBy(mgr).
		For(&webv1.WebServer{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
