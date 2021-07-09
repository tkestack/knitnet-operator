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

package ensures

const (
	SubmarinerOperatorNamespace = "submariner-operator"
	SubmarinerOperatorImage     = "quay.io/submariner/submariner-operator:0.9.1"

	SubmarinerBrokerName      = "submariner-broker"
	SubmarinerBrokerNamespace = "submariner-k8s-broker"

	// SubmarinerBrokerInfo represents the broker info configmap name
	SubmarinerBrokerInfo = "submariner-broker-info"

	//FabricNameLabel is the label used to label the resource managed by fabric
	FabricNameLabel = "operator.tkestack.io/fabric-name"

	//FabricNamespaceLabel is the label used to label the resource managed by fabric
	FabricNamespaceLabel = "operator.tkestack.io/fabric-namespace"
)
