package controller

import (
	corev1 "k8s.io/api/core/v1"

	webv1 "github.com/XiongGG1989/web-operator/api/v1"
)

type webServerConfig struct {
	targetNamespace string
	replicas        int32
	image           string
	port            int32
	serviceType     corev1.ServiceType
	labels          map[string]string
}

func buildWebServerConfig(ws *webv1.WebServer) webServerConfig {
	targetNamespace := ws.Namespace
	if ws.Spec.TargetNamespace != "" {
		targetNamespace = ws.Spec.TargetNamespace
	}

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

	serviceType := corev1.ServiceTypeClusterIP
	switch ws.Spec.ServiceType {
	case "NodePort":
		serviceType = corev1.ServiceTypeNodePort
	case "LoadBalancer":
		serviceType = corev1.ServiceTypeLoadBalancer
	}

	return webServerConfig{
		targetNamespace: targetNamespace,
		replicas:        replicas,
		image:           image,
		port:            port,
		serviceType:     serviceType,
		labels: map[string]string{
			"app.kubernetes.io/name":     "webserver",
			"app.kubernetes.io/instance": ws.Name,
		},
	}
}
