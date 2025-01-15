package v1alpha1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CNFSpec defines the desired state of CNF
type CNFSpec struct {
    HelmChartURL string `json:"helmChartURL"`
}

// CNFStatus defines the observed state of CNF
type CNFStatus struct {
    Deployed bool `json:"deployed"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CNF is the Schema for the cnfs API
type CNF struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   CNFSpec   `json:"spec,omitempty"`
    Status CNFStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CNFList contains a list of CNF
type CNFList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []CNF `json:"items"`
}

func init() {
    SchemeBuilder.Register(&CNF{}, &CNFList{})
}
