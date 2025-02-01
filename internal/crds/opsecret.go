//+kubebuilder:object:generate=true
//+groupName=crds.driscoll.co
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

// OPSecret is the intention to create a secret from a 1Password item
type OPSecret struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	metav1.TypeMeta   `json:",inline"`
	Spec              OPSecretSpec   `json:"spec,omitempty"`
	Status            OPSecretStatus `json:"status,omitempty"`
}

// OPSecretStatus defines the state of a secret as it is created
type OPSecretStatus struct {
	Phase       string             `json:"phase,omitempty"`
	Conditions  []metav1.Condition `json:"conditions,omitempty"`
	Events      []Event            `json:"events,omitempty"`
	LastUpdated metav1.Time        `json:"last-updated"`
}

type Event struct {
	Timestamp metav1.Time `json:"timestamp"`
	Type      string      `json:"type"` // e.g., "Created", "Updated", "Failed"
	Reason    string      `json:"reason"`
	Message   string      `json:"message"`
}

// OPSecretSpec contains instructions on how to source and create a secret
type OPSecretSpec struct {
	Source SourceConfig `json:"source"`
	Secret SecretConfig `json:"secret"`
}

// SourceConfig defines the location within 1Password where the information can be found
type SourceConfig struct {
	Vault   string `json:"vault"`
	Item    string `json:"item"`
	Section string `json:"section"`
}

type KeyMapping struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// SecretConfig defines the location within Kubernetes where the secret should be created
type SecretConfig struct {
	// The name of the secret
	Name string `json:"name"`
	// Check this secret every N seconds in 1Password and update the secret if anything changes
	RefreshSeconds int `json:"refresh-seconds"`
	// +kubebuilder:validation:Enum=basic;docker
	// Type of secret. Leave unpopulated for a standard secret. Choose docker for a secret which can be used to pull images from a registry.
	// If using 'docker' as the type you don't need to specify any keys as this will be done for you
	// Possible types:
	//   * basic - Standard secret
	//   * docker - Secret used for pulling images from a docker registry
	SecretType string `json:"secret-type"`
	// Keys maps individual 1Password section keys to data items within a secret
	// This does not need to be populated for Docker secret types as this will be calculated by the operator
	Keys []KeyMapping `json:"keys,omitempty"`
}

//go:generate controller-gen object crd paths=./... output:crd:dir=../../cmd/build/helm/crds

// +kubebuilder:object:root=true
type OPSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OPSecret `json:"items"`
}
