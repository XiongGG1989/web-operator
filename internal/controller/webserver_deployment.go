package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	webv1 "github.com/XiongGG1989/web-operator/api/v1"
)

func (r *WebServerReconciler) reconcileDeployment(
	ctx context.Context,
	ws *webv1.WebServer,
	cfg webServerConfig,
) (controllerutil.OperationResult, bool, error) {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: ws.Name, Namespace: cfg.namespace},
	}

	var changed bool
	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, deploy, func() error {
		if deploy.ResourceVersion != "" {
			changed = deploymentSpecChanged(deploy, cfg)
		}

		deploy.Labels = cfg.labels
		deploy.Spec = appsv1.DeploymentSpec{
			Replicas: &cfg.replicas,
			Selector: &metav1.LabelSelector{MatchLabels: cfg.labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: cfg.labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:      "web",
						Image:     cfg.image,
						Resources: ws.Spec.Resources,
						Ports:     []corev1.ContainerPort{{ContainerPort: cfg.port}},
					}},
				},
			},
		}

		return controllerutil.SetControllerReference(ws, deploy, r.Scheme)
	})

	return result, changed, err
}

func deploymentSpecChanged(deploy *appsv1.Deployment, cfg webServerConfig) bool {
	if deploy.Spec.Replicas != nil && *deploy.Spec.Replicas != cfg.replicas {
		return true
	}
	if len(deploy.Spec.Template.Spec.Containers) == 0 {
		return false
	}

	container := deploy.Spec.Template.Spec.Containers[0]
	if container.Image != cfg.image {
		return true
	}

	return len(container.Ports) > 0 && container.Ports[0].ContainerPort != cfg.port
}

func (r *WebServerReconciler) logDeploymentResult(
	log interface {
		Info(string, ...any)
	},
	ws *webv1.WebServer,
	cfg webServerConfig,
	result controllerutil.OperationResult,
	changed bool,
) {
	if result == controllerutil.OperationResultCreated {
		log.Info("Deployment created", "namespace", cfg.namespace, "name", ws.Name, "replicas", cfg.replicas, "image", cfg.image)
	}
	if result == controllerutil.OperationResultUpdated && changed {
		log.Info("Deployment updated", "namespace", cfg.namespace, "name", ws.Name)
	}
}
