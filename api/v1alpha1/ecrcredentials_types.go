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

// ECRCredentialsSpec defines the desired state of ECRCredentials
type ECRCredentialsSpec struct {
	//+kubebuilder:validation:Required
	AWSAccessKeyID string `json:"aws_access_key_id"`

	//+kubebuilder:validation:Required
	AWSSecretAccessKey string `json:"aws_secret_access_key"`

	//+kubebuilder:validation:Required
	AWSRegion string `json:"aws_region"`
}

// ECRCredentialsStatus defines the observed state of ECRCredentials
type ECRCredentialsStatus struct {
	//+kubebuilder:validation:Optional
	Phase ECRCredentialsPhase `json:"phase,omitempty"`
}

type ECRCredentialsPhase string

var (
	ECRCredentialsProvisioning ECRCredentialsPhase = "Provisioning"
	ECRCredentialsActive       ECRCredentialsPhase = "Active"
	ECRCredentialsTerminating  ECRCredentialsPhase = "Terminanting"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`

// ECRCredentials is the Schema for the ecrcredentials API
type ECRCredentials struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ECRCredentialsSpec   `json:"spec,omitempty"`
	Status ECRCredentialsStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ECRCredentialsList contains a list of ECRCredentials
type ECRCredentialsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ECRCredentials `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ECRCredentials{}, &ECRCredentialsList{})
}
