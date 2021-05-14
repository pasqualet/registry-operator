/*
Copyright 2021 AstroKube.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DockerioCredentialsSpec defines the desired state of DockerioCredentials
type DockerioCredentialsSpec struct {
	//+kubebuilder:validation:Required
	User string `json:"user"`

	//+kubebuilder:validation:Required
	Password string `json:"password"`
}

// DockerioCredentialsStatus defines the observed state of DockerioCredentials
type DockerioCredentialsStatus struct {
	//+kubebuilder:validation:Optional
	Phase DockerioCredentialsPhase `json:"phase,omitempty"`
}

type DockerioCredentialsPhase string

var (
	DockerioCredentialsProvisioning DockerioCredentialsPhase = "Provisioning"
	DockerioCredentialsActive       DockerioCredentialsPhase = "Active"
	DockerioCredentialsTerminating  DockerioCredentialsPhase = "Terminanting"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`

// DockerioCredentials is the Schema for the dockeriocredentials API
type DockerioCredentials struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DockerioCredentialsSpec   `json:"spec,omitempty"`
	Status DockerioCredentialsStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DockerioCredentialsList contains a list of DockerioCredentials
type DockerioCredentialsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DockerioCredentials `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DockerioCredentials{}, &DockerioCredentialsList{})
}
