package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WordpressSpec defines the desired state of Wordpress
type WordpressSpec struct {
	Password string `json:"sqlRootPassword"`
}

// WordpressStatus defines the observed state of Wordpress
type WordpressStatus struct {
	// Requirements do not require the state of the resource to be tracked.
	// TODO: Track Pods, PVC, and Secret.
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Wordpress is the Schema for the wordpresses API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=wordpresses,scope=Namespaced
type Wordpress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WordpressSpec   `json:"spec,omitempty"`
	Status WordpressStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WordpressList contains a list of Wordpress
type WordpressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Wordpress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Wordpress{}, &WordpressList{})
}
