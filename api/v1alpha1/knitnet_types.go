/*
Copyright 2021.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KnitnetSpec defines the desired state of Knitnet
type KnitnetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Action represents deploy broker or join cluster to broker
	// +optional
	// +kubebuilder:default=broker
	// +kubebuilder:validation:Enum=broker;join;all
	Action string `json:"action,omitempty"`

	// BrokerConfig represents the broker cluster configuration of the Submariner.
	// +optional
	BrokerConfig `json:"brokerConfig,omitempty"`

	// JoinConfig represents the managed cluster join configuration of the Submariner.
	// +optional
	JoinConfig `json:"joinConfig,omitempty"`

	// CloudPrepareConfig represents the prepare config for the cloud vendor.
	// +optional
	CloudPrepareConfig `json:"cloudPrepareConfig,omitempty"`
}

// KnitnetStatus defines the observed state of Knitnet
type KnitnetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Phase is the knitnet operator running phase.
	// +optional
	Phase Phase `json:"phase,omitempty"`
}

const (
	PhaseRunning Phase = "Running"
	PhaseFailed  Phase = "Failed"
)

// Phase is the phase of the installation.
type Phase string

type BrokerConfig struct {
	// ConnectivityEnabled represents enable/disable multi-cluster pod connectivity.
	// +optional
	// +kubebuilder:default=true
	ConnectivityEnabled bool `json:"connectivityEnabled,omitempty"`
	// ServiceDiscoveryEnabled represents enable/disable multi-cluster service discovery.
	// +optional
	// +kubebuilder:default=false
	ServiceDiscoveryEnabled bool `json:"serviceDiscoveryEnabled,omitempty"`
	// GlobalnetEnable represents enable/disable overlapping CIDRs in connecting clusters (default disabled).
	// +optional
	// +kubebuilder:default=false
	GlobalnetEnable bool `json:"globalnetEnable,omitempty"`
	// GlobalnetCIDRRange represents global CIDR supernet range for allocating global CIDRs to each cluster.
	// +optional
	// +kubebuilder:default="242.0.0.0/8"
	GlobalnetCIDRRange string `json:"globalnetCIDRRange,omitempty"`
	// DefaultGlobalnetClusterSize represents default cluster size for global CIDR allocated to each cluster (amount of global IPs).
	// +optional
	// +kubebuilder:default=65336
	DefaultGlobalnetClusterSize uint `json:"defaultGlobalnetClusterSize,omitempty"`
	// DefaultCustomDomains represents list of domains to use for multicluster service discovery.
	// +optional
	DefaultCustomDomains []string `json:"defaultCustomDomains,omitempty"`
}

type JoinConfig struct {
	// ClusterID used to identify the tunnels.
	ClusterID string `json:"clusterID"`
	// ServiceCIDR represents service CIDR.
	// +optional
	ServiceCIDR string `json:"serviceCIDR,omitempty"`
	// ClusterCIDR represents cluster CIDR.
	// +optional
	ClusterCIDR string `json:"clusterCIDR,omitempty"`
	// GlobalCIDR represents global CIDR to be allocated to the cluster.
	// +optional
	GlobalnetCIDR string `json:"globalnetCIDR,omitempty"`
	// Repository represents image repository.
	// +optional
	Repository string `json:"repository,omitempty"`
	// ImageVersion represents image version.
	// +optional
	ImageVersion string `json:"imageVersion,omitempty"`
	// NattPort represents IPsec NAT-T port (default 4500).
	// +optional
	// +kubebuilder:default=4500
	NattPort int `json:"nattPort,omitempty"`
	// IkePort represents IPsec IKE port (default 500).
	// +optional
	// +kubebuilder:default=500
	IkePort int `json:"ikePort,omitempty"`
	// PreferredServer represents enable/disable this cluster as a preferred server for data-plane connections.
	// +optional
	// +kubebuilder:default=false
	PreferredServer bool `json:"preferredServer,omitempty"`
	// ForceUDPEncaps represents force UDP encapsulation for IPSec.
	// +optional
	// +kubebuilder:default=false
	ForceUDPEncaps bool `json:"forceUDPEncaps,omitempty"`
	// NatTraversal represents enable NAT traversal for IPsec
	// +optional
	// +kubebuilder:default=true
	NatTraversal bool `json:"natTraversal,omitempty"`
	// GlobalnetEnabled represents enable/disable Globalnet for this cluster.
	// +optional
	// +kubebuilder:default=true
	GlobalnetEnabled bool `json:"globalnetEnabled,omitempty"`
	// IpsecDebug represents enable/disable IPsec debugging (verbose logging).
	// +optional
	// +kubebuilder:default=false
	IpsecDebug bool `json:"ipsecDebug,omitempty"`
	// SubmarinerDebug represents enable/disable submariner pod debugging (verbose logging in the deployed pods).
	// +optional
	// +kubebuilder:default=false
	SubmarinerDebug bool `json:"submarinerDebug,omitempty"`
	// LabelGateway represents enable/disable label gateways.
	// +optional
	// +kubebuilder:default=true
	LabelGateway bool `json:"labelGateway,omitempty"`
	// LoadBalancerEnabled represents enable/disable automatic LoadBalancer in front of the gateways.
	// +optional
	// +kubebuilder:default=false
	LoadBalancerEnabled bool `json:"loadBalancerEnabled,omitempty"`
	// CableDriver represents cable driver implementation.
	// +optional
	CableDriver string `json:"cableDriver,omitempty"`
	// GlobalnetClusterSize represents cluster size for GlobalCIDR allocated to this cluster (amount of global IPs).
	// +optional
	// +kubebuilder:default=0
	GlobalnetClusterSize uint `json:"globalnetClusterSize,omitempty"`
	// CustomDomains represents list of domains to use for multicluster service discovery.
	// +optional
	CustomDomains []string `json:"customDomains,omitempty"`
	// ImageOverrideArr represents override component image.
	// +optional
	ImageOverrideArr []string `json:"imageOverrideArr,omitempty"`
	// HealthCheckEnable represents enable/disable gateway health check.
	// +optional
	// +kubebuilder:default=true
	HealthCheckEnable bool `json:"healthCheckEnable,omitempty"`
	// HealthCheckInterval represents interval in seconds between health check packets.
	// +optional
	// +kubebuilder:default=1
	HealthCheckInterval uint64 `json:"healthCheckInterval,omitempty"`
	// HealthCheckMaxPacketLossCount represents maximum number of packets lost before the connection is marked as down.
	// +optional
	// +kubebuilder:default=5
	HealthCheckMaxPacketLossCount uint64 `json:"healthCheckMaxPacketLossCount,omitempty"`
	// CorednsCustomConfigMap represents name of the custom CoreDNS configmap to configure forwarding to lighthouse. It should be in
	// <namespace>/<name> format where <namespace> is optional and defaults to kube-system
	// +optional
	CorednsCustomConfigMap string `json:"corednsCustomConfigMap,omitempty"`
}

type CloudPrepareConfig struct {
	// CredentialsSecret is a reference to the secret with a certain cloud platform
	// credentials, the supported platform includes AWS, GCP, Azure, ROKS and OSD.
	// The knitnet-operator will use these credentials to prepare Submariner cluster
	// environment. If the submariner cluster environment requires knitnet-operator
	// preparation, this field should be specified.
	// +optional
	CredentialsSecret *corev1.LocalObjectReference `json:"credentialsSecret,omitempty"`

	// Infra ID
	InfraID string `json:"infraID,omitempty"`
	// Regio
	Region string `json:"region,omitempty"`

	// AWS specific cloud prepare setup
	AWS `json:"aws,omitempty"`
}

type AWS struct {
	// GatewayInstance represents type of gateways instance machine (default "m5n.large")
	// +optional
	// +kubebuilder:default=m5n.large
	GatewayInstance string `json:"gatewayInstance,omitempty"`

	// Gateways represents the count of worker nodes that will be used to deploy the Submariner gateway
	// component on the managed cluster.
	// +optional
	// +kubebuilder:default=1
	Gateways int `json:"gateways,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:path=knitnets,shortName=fb,scope=Namespaced
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=.metadata.creationTimestamp
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=.status.phase,description="Current Cluster Phase"
// +kubebuilder:printcolumn:name="Created At",type=string,JSONPath=.metadata.creationTimestamp
// Knitnet is the Schema for the knitnets API
type Knitnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KnitnetSpec   `json:"spec,omitempty"`
	Status KnitnetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KnitnetList contains a list of Knitnet
type KnitnetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Knitnet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Knitnet{}, &KnitnetList{})
}
