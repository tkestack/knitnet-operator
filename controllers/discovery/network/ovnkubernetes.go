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
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	constants "github.com/tkestack/cluster-fabric-operator/controllers/discovery"
)

const (
	ovnKubeService     = "ovnkube-db"
	OvnNBDB            = "OVN_NBDB"
	OvnSBDB            = "OVN_SBDB"
	OvnNBDBDefaultPort = 6641
	OvnSBDBDefaultPort = 6642
	OvnKubernetes      = "OVNKubernetes"
)

func discoverOvnKubernetesNetwork(c client.Client) (*ClusterNetwork, error) {
	ovnDBPod, err := findPod(c, "name=ovnkube-db")

	if err != nil || ovnDBPod == nil {
		return nil, err
	}
	svc := &corev1.Service{}
	svcKey := types.NamespacedName{Name: ovnKubeService, Namespace: ovnDBPod.Namespace}
	if err := c.Get(context.TODO(), svcKey, svc); err != nil {
		return nil, fmt.Errorf("error finding %q service in %q namespace", ovnKubeService, ovnDBPod.Namespace)
	}

	dbConnectionProtocol := "tcp"

	for _, container := range ovnDBPod.Spec.Containers {
		for _, envVar := range container.Env {
			if envVar.Name == "OVN_SSL_ENABLE" {
				if strings.ToUpper(envVar.Value) != "NO" {
					dbConnectionProtocol = "ssl"
				}
			}
		}
	}

	clusterNetwork := &ClusterNetwork{
		NetworkPlugin: constants.NetworkPluginOVNKubernetes,
		PluginSettings: map[string]string{
			OvnNBDB: fmt.Sprintf("%s:%s.%s:%d", dbConnectionProtocol, ovnKubeService, ovnDBPod.Namespace, OvnNBDBDefaultPort),
			OvnSBDB: fmt.Sprintf("%s:%s.%s:%d", dbConnectionProtocol, ovnKubeService, ovnDBPod.Namespace, OvnSBDBDefaultPort),
		},
	}

	// If the cluster/service CIDRs weren't found we leave it to the generic functions to figure out later
	cm := &corev1.ConfigMap{}
	cmKey := types.NamespacedName{Name: "ovn-config", Namespace: ovnDBPod.Namespace}
	if err := c.Get(context.TODO(), cmKey, cm); err == nil {
		if netCidr, ok := cm.Data["net_cidr"]; ok {
			clusterNetwork.PodCIDRs = []string{netCidr}
		}

		if svcCidr, ok := cm.Data["svc_cidr"]; ok {
			clusterNetwork.ServiceCIDRs = []string{svcCidr}
		}
	}

	return clusterNetwork, nil
}
