// Copyright 2020 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"github.com/kiegroup/kogito-operator/apis"
	"github.com/kiegroup/kogito-operator/apis/app/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +genclient
// +kubebuilder:resource:path=kogitosupportingservices,scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas",description="Number of replicas set for this service"
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".status.image",description="Base image for this service"
// +kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".status.externalURI",description="External URI to access this service"
// +kubebuilder:printcolumn:name="Service Type",type="string",JSONPath=".spec.serviceType",description="Supporting Service Type"
// +operator-sdk:csv:customresourcedefinitions:displayName="Kogito Supporting Service"
// +operator-sdk:csv:customresourcedefinitions:resources={{Deployment,apps/v1,"A Kubernetes Deployment"}}
// +operator-sdk:csv:customresourcedefinitions:resources={{Service,v1,"A Kubernetes Service"}}
// +operator-sdk:csv:customresourcedefinitions:resources={{ImageStream,image.openshift.io/v1,"A Openshift ImageStream"}}
// +operator-sdk:csv:customresourcedefinitions:resources={{Route,route.openshift.io/v1,"A Openshift Route"}}

// KogitoSupportingService deploys the Supporting service in the given namespace.
type KogitoSupportingService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   v1beta1.KogitoSupportingServiceSpec   `json:"spec,omitempty"`
	Status v1beta1.KogitoSupportingServiceStatus `json:"status,omitempty"`
}

// GetSpec ...
func (k *KogitoSupportingService) GetSpec() api.KogitoServiceSpecInterface {
	return &k.Spec
}

// GetStatus ...
func (k *KogitoSupportingService) GetStatus() api.KogitoServiceStatusInterface {
	return &k.Status
}

// GetSupportingServiceSpec ...
func (k *KogitoSupportingService) GetSupportingServiceSpec() api.KogitoSupportingServiceSpecInterface {
	return &k.Spec
}

// GetSupportingServiceStatus ...
func (k *KogitoSupportingService) GetSupportingServiceStatus() api.KogitoSupportingServiceStatusInterface {
	return &k.Status
}

// +kubebuilder:object:root=true

// KogitoSupportingServiceList contains a list of KogitoSupportingService.
type KogitoSupportingServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KogitoSupportingService `json:"items"`
}

// GetItems ...
func (k *KogitoSupportingServiceList) GetItems() []api.KogitoSupportingServiceInterface {
	models := make([]api.KogitoSupportingServiceInterface, len(k.Items))
	for i, v := range k.Items {
		item := v
		models[i] = &item
	}
	return models
}

func init() {
	SchemeBuilder.Register(&KogitoSupportingService{}, &KogitoSupportingServiceList{})
}
