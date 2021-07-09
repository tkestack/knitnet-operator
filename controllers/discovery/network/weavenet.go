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

package network

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	constants "github.com/tkestack/cluster-fabric-operator/controllers/discovery"
)

func discoverWeaveNetwork(c client.Client) (*ClusterNetwork, error) {
	weaveNetPod, err := findPod(c, "name=weave-net")

	if err != nil || weaveNetPod == nil {
		return nil, err
	}

	var clusterNetwork *ClusterNetwork

	for _, container := range weaveNetPod.Spec.Containers {
		for _, envVar := range container.Env {
			if envVar.Name == "IPALLOC_RANGE" {
				clusterNetwork = &ClusterNetwork{
					PodCIDRs:      []string{envVar.Value},
					NetworkPlugin: constants.NetworkPluginWeaveNet,
				}
				break
			}
		}
	}

	if clusterNetwork == nil {
		return nil, nil
	}

	clusterIPRange, err := findClusterIPRange(c)
	if err == nil && clusterIPRange != "" {
		clusterNetwork.ServiceCIDRs = []string{clusterIPRange}
	}

	return clusterNetwork, nil
}
