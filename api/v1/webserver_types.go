/*
Copyright 2026 xiongming.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WebServerSpec defines the desired state of WebServer
type WebServerSpec struct {
	// 副本数
	// +optional
	// +kubebuilder:validation:Minimum=0
	Replicas *int32 `json:"replicas,omitempty"`

	// 镜像名称
	Image string `json:"image,omitempty"`

	// 端口号
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port,omitempty"`

	// Service 类型
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	// +optional
	ServiceType string `json:"serviceType,omitempty"`

	// 资源限制
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// 目标部署 namespace（默认为 CR 所在 namespace）
	// +optional
	TargetNamespace string `json:"targetNamespace,omitempty"`
}

// WebServerStatus defines the observed state of WebServer.
type WebServerStatus struct {
	// 可用副本数
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
	// +listType=map

	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// WebServer is the Schema for the webservers API
type WebServer struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of WebServer
	// +required
	Spec WebServerSpec `json:"spec"`

	// status defines the observed state of WebServer
	// +optional
	Status WebServerStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// WebServerList contains a list of WebServer
type WebServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []WebServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WebServer{}, &WebServerList{})
}
