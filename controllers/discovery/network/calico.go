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
	"sigs.k8s.io/controller-runtime/pkg/client"

	constants "github.com/tkestack/knitnet-operator/controllers/discovery"
)

func discoverCalicoNetwork(c client.Client) (*ClusterNetwork, error) {
	cmList := &v1.ConfigMapList{}
	err := c.List(context.TODO(), cmList)
	if err != nil {
		return nil, err
	}

	findCalicoConfigMap := false
	for _, cm := range cmList.Items {
		if cm.Name == "calico-config" {
			findCalicoConfigMap = true
			break
		}
	}

	if !findCalicoConfigMap {
		return nil, nil
	}

	clusterNetwork, err := discoverNetwork(c)
	if err != nil {
		return nil, err
	}

	if clusterNetwork != nil {
		clusterNetwork.NetworkPlugin = constants.NetworkPluginCalico
		return clusterNetwork, nil
	}

	return nil, nil
}
