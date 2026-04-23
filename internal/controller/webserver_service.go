package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	webv1 "github.com/XiongGG1989/web-operator/api/v1"
)

func (r *WebServerReconciler) reconcileService(
	ctx context.Context,
	ws *webv1.WebServer,
	cfg webServerConfig,
) (controllerutil.OperationResult, bool, error) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: ws.Name, Namespace: cfg.namespace},
	}

	var changed bool
	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		if svc.ResourceVersion != "" {
			changed = serviceSpecChanged(svc, cfg)
		}

		svc.Labels = cfg.labels
		svc.Spec = corev1.ServiceSpec{
			Type:     cfg.serviceType,
			Selector: cfg.labels,
			Ports: []corev1.ServicePort{{
				Port:       cfg.port,
				TargetPort: intstr.FromInt(int(cfg.port)),
				Protocol:   corev1.ProtocolTCP,
			}},
		}

		return controllerutil.SetControllerReference(ws, svc, r.Scheme)
	})

	return result, changed, err
}

func serviceSpecChanged(svc *corev1.Service, cfg webServerConfig) bool {
	if len(svc.Spec.Ports) > 0 && svc.Spec.Ports[0].Port != cfg.port {
		return true
	}

	return svc.Spec.Type != cfg.serviceType
}

func (r *WebServerReconciler) logServiceResult(
	log interface {
		Info(string, ...interface{})
	},
	ws *webv1.WebServer,
	cfg webServerConfig,
	result controllerutil.OperationResult,
	changed bool,
) {
	if result == controllerutil.OperationResultCreated {
		log.Info("Service created", "namespace", cfg.namespace, "name", ws.Name, "type", cfg.serviceType, "port", cfg.port)
	}
	if result == controllerutil.OperationResultUpdated && changed {
		log.Info("Service updated", "namespace", cfg.namespace, "name", ws.Name)
	}
}
