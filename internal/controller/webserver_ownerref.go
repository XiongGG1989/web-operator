package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	webv1 "github.com/XiongGG1989/web-operator/api/v1"
)

func (r *WebServerReconciler) setOwnerReferenceIfAllowed(ws *webv1.WebServer, obj metav1.Object) error {
	if ws.Namespace != obj.GetNamespace() {
		controllerutil.RemoveControllerReference(ws, obj, r.Scheme)
		return nil
	}

	return controllerutil.SetControllerReference(ws, obj, r.Scheme)
}
