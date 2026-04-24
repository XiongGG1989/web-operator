package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	webv1 "github.com/XiongGG1989/web-operator/api/v1"
)

func (r *WebServerReconciler) reconcileStatus(
	ctx context.Context,
	log interface {
		Info(string, ...any)
	},
	ws *webv1.WebServer,
	cfg webServerConfig,
) error {
	currentDeploy := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: ws.Name, Namespace: cfg.namespace}, currentDeploy); err != nil {
		setCondition(&ws.Status, metav1.Condition{
			Type:    "Degraded",
			Status:  metav1.ConditionTrue,
			Reason:  "DeploymentNotFound",
			Message: err.Error(),
		})
	} else {
		if currentDeploy.Status.AvailableReplicas != ws.Status.AvailableReplicas {
			log.Info(
				"Pod replicas changed",
				"namespace", cfg.namespace,
				"name", ws.Name,
				"available", currentDeploy.Status.AvailableReplicas,
				"desired", *currentDeploy.Spec.Replicas,
				"ready", currentDeploy.Status.ReadyReplicas,
				"updated", currentDeploy.Status.UpdatedReplicas,
			)
		}

		ws.Status.AvailableReplicas = currentDeploy.Status.AvailableReplicas
		setCondition(&ws.Status, availabilityCondition(currentDeploy))
		setCondition(&ws.Status, progressingCondition(currentDeploy))
	}

	return r.Status().Update(ctx, ws)
}

func availabilityCondition(deploy *appsv1.Deployment) metav1.Condition {
	available := deploy.Status.AvailableReplicas > 0

	return metav1.Condition{
		Type:    "Available",
		Status:  boolToCondition(available),
		Reason:  map[bool]string{true: "DeploymentReady", false: "NoAvailableReplicas"}[available],
		Message: map[bool]string{true: "Deployment has available replicas", false: "No available replicas"}[available],
	}
}

func progressingCondition(deploy *appsv1.Deployment) metav1.Condition {
	updating := deploy.Spec.Replicas != nil && deploy.Status.UpdatedReplicas != *deploy.Spec.Replicas

	return metav1.Condition{
		Type:    "Progressing",
		Status:  boolToCondition(updating),
		Reason:  map[bool]string{true: "Updating", false: "UpToDate"}[updating],
		Message: map[bool]string{true: "Deployment is updating", false: "Deployment is up-to-date"}[updating],
	}
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
