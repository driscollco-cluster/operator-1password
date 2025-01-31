//+kubebuilder:object:generate=true
//+groupName=crds.driscollco
//+versionName=v1

package crds

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:type=customType
//+kubebuilder:printcolumn:name="Vault",type=string,JSONPath=".spec.source-vault",description="The vault the secret is sourced from"
//+kubebuilder:printcolumn:name="Item",type=string,JSONPath=".spec.source-item",description="The item the secret is sourced from"
//+kubebuilder:printcolumn:name="Section",type=string,JSONPath=".spec.source-section",description="The item the secret is sourced from"
//+kubebuilder:printcolumn:name="Key",type=string,JSONPath=".spec.source-key",description="The name of the subitem which contains the secret"

// ExternalSecret is the intention to create a secret from a 1Password item
type ExternalSecret struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	metav1.TypeMeta   `json:",inline"`
	Spec              ExternalSecretSpec   `json:"spec,omitempty"`
	Status            ExternalSecretStatus `json:"status,omitempty"`
}

// ExternalSecretStatus defines the state of a secret as it is created
type ExternalSecretStatus struct {
	Phase      string             `json:"phase,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	Events     []Event            `json:"events"`
}

type Event struct {
	Timestamp metav1.Time `json:"timestamp"`
	Type      string      `json:"type"` // e.g., "Created", "Updated", "Failed"
	Reason    string      `json:"reason"`
	Message   string      `json:"message"`
}

// ExternalSecretSpec contains instructions on how to source and create a secret
type ExternalSecretSpec struct {
	SourceVault           string `json:"source-vault"`
	SourceItem            string `json:"source-item"`
	SourceSection         string `json:"source-section"`
	SourceKey             string `json:"source-key"`
	DestinationNamespace  string `json:"destination-namespace"`
	DestinationName       string `json:"destination-name"`
	DestinationSecretType string `json:"destination-secret-type"`
}

//go:generate controller-gen object crd paths=./... output:crd:dir=../../cmd/build/helm/crds

// +kubebuilder:object:root=true
type ExternalSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExternalSecret `json:"items"`
}
