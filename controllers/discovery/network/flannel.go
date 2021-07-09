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
	"context"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	constants "github.com/tkestack/cluster-fabric-operator/controllers/discovery"
)

func discoverFlannelNetwork(c client.Client) (*ClusterNetwork, error) {
	cm := &v1.ConfigMap{}
	cmKey := types.NamespacedName{Name: "kube-flannel-cfg", Namespace: "kube-system"}
	err := c.Get(context.TODO(), cmKey, cm)

	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		klog.Errorf("error obtaining the \"kube-flannel-cfg\" ConfigMapz: %v", err)
		return nil, err
	}

	podCIDR := extractPodCIDRFromNetConfigJSON(cm)

	if podCIDR == nil {
		return nil, nil
	}

	clusterNetwork := &ClusterNetwork{
		NetworkPlugin: constants.NetworkPluginFlannel,
		PodCIDRs:      []string{*podCIDR},
	}

	// Try to detect the service CIDRs using the generic functions
	clusterIPRange, err := findClusterIPRange(c)
	if err != nil {
		return nil, err
	}

	if clusterIPRange != "" {
		clusterNetwork.ServiceCIDRs = []string{clusterIPRange}
	}

	return clusterNetwork, nil
}
