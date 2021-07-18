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

package checker

import (
	"context"

	submarinerv1 "github.com/submariner-io/submariner/pkg/apis/submariner.io/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	netconsts "github.com/tkestack/knitnet-operator/controllers/discovery"
	"github.com/tkestack/knitnet-operator/controllers/ensures/broker"
	"github.com/tkestack/knitnet-operator/controllers/ensures/common/ippools"
)

func EnsureCalico(c client.Client, config *rest.Config, currentClusterID string, clusterInfos *[]broker.ClusterInfo) error {
	if err := CreateOrUpdateIPPools(c, config, currentClusterID, clusterInfos); err != nil {
		return err
	}
	return nil
}

func CreateOrUpdateIPPools(c client.Client, config *rest.Config, currentClusterID string, clusterInfos *[]broker.ClusterInfo) error {
	klog.Infof("Creating IPPools")
	clusters, err := GetClusters(c)
	if err != nil {
		return err
	}
	for _, clusterInfo := range *clusterInfos {
		if clusterInfo.ClusterID == currentClusterID || clusterInfo.NetworkPlugin != netconsts.NetworkPluginCalico {
			continue
		}
		cluster := GetClusterWithID(clusterInfo.ClusterID, clusters)
		if err := ippools.EnsureIPPool(config, cluster.Spec.ClusterID+"-pod-cidr", cluster.Spec.ClusterCIDR[0]); err != nil {
			return err
		}
		if err := ippools.EnsureIPPool(config, cluster.Spec.ClusterID+"-svc-cidr", cluster.Spec.ServiceCIDR[0]); err != nil {
			return err
		}
	}
	return nil
}

func GetClusterWithID(ID string, clusters *submarinerv1.ClusterList) *submarinerv1.Cluster {
	for _, cluster := range clusters.Items {
		if cluster.Spec.ClusterID == ID {
			return &cluster
		}
	}
	return nil
}

func GetClusters(c client.Client) (*submarinerv1.ClusterList, error) {
	clusters := &submarinerv1.ClusterList{}
	if err := c.List(context.TODO(), clusters); err != nil {
		klog.Errorf("Failed to list Cluster: %v", err)
		return nil, err
	}
	return clusters, nil
}
