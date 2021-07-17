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
	netconsts "github.com/tkestack/knitnet-operator/controllers/discovery"
	"github.com/tkestack/knitnet-operator/controllers/embeddedyamls"
	"github.com/tkestack/knitnet-operator/controllers/ensures/broker"
	"github.com/tkestack/knitnet-operator/controllers/ensures/common/configmaps"
	"github.com/tkestack/knitnet-operator/controllers/ensures/common/daemonsets"
	"github.com/tkestack/knitnet-operator/controllers/ensures/common/deployments"
	"github.com/tkestack/knitnet-operator/controllers/ensures/common/ippools"
	"github.com/tkestack/knitnet-operator/controllers/ensures/common/poddisruptionbudgets"
	"github.com/tkestack/knitnet-operator/controllers/ensures/common/serviceaccount"
	"github.com/tkestack/knitnet-operator/controllers/utils"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var kddCrds = []string{
	embeddedyamls.Manifests_fix_calico_crds_crd_projectcalico_org_bgpconfigurations_yaml,
	embeddedyamls.Manifests_fix_calico_crds_crd_projectcalico_org_bgppeers_yaml,
	embeddedyamls.Manifests_fix_calico_crds_crd_projectcalico_org_blockaffinities_yaml,
	embeddedyamls.Manifests_fix_calico_crds_crd_projectcalico_org_clusterinformations_yaml,
	embeddedyamls.Manifests_fix_calico_crds_crd_projectcalico_org_felixconfigurations_yaml,
	embeddedyamls.Manifests_fix_calico_crds_crd_projectcalico_org_globalnetworkpolicies_yaml,
	embeddedyamls.Manifests_fix_calico_crds_crd_projectcalico_org_globalnetworksets_yaml,
	embeddedyamls.Manifests_fix_calico_crds_crd_projectcalico_org_hostendpoints_yaml,
	embeddedyamls.Manifests_fix_calico_crds_crd_projectcalico_org_ipamblocks_yaml,
	embeddedyamls.Manifests_fix_calico_crds_crd_projectcalico_org_ipamconfigs_yaml,
	embeddedyamls.Manifests_fix_calico_crds_crd_projectcalico_org_ipamhandles_yaml,
	embeddedyamls.Manifests_fix_calico_crds_crd_projectcalico_org_ippools_yaml,
	embeddedyamls.Manifests_fix_calico_crds_crd_projectcalico_org_kubecontrollersconfigurations_yaml,
	embeddedyamls.Manifests_fix_calico_crds_crd_projectcalico_org_networkpolicies_yaml,
	embeddedyamls.Manifests_fix_calico_crds_crd_projectcalico_org_networksets_yaml,
}

func EnsureCalico(c client.Client) error {
	if err := CreateOrUpdateKddCRD(c); err != nil {
		return err
	}
	if err := CreateOrUpdateServiceAccount(c); err != nil {
		return err
	}
	if err := CreateOrUpdateClusterRole(c); err != nil {
		return err
	}
	if err := CreateOrUpdateClusterRoleBinding(c); err != nil {
		return err
	}
	if err := CreateOrUpdateDeployment(c); err != nil {
		return err
	}
	if err := CreateOrUpdateDaemonSet(c); err != nil {
		return err
	}
	if err := CreateOrUpdateConfigMap(c); err != nil {
		return err
	}
	if err := CreateOrUpdatePodDisruptionBudget(c); err != nil {
		return err
	}
	return nil
}

func CreateOrUpdateIPPools(c client.Client, config *rest.Config, currentClusterID string, clusterInfos *[]broker.ClusterInfo) error {
	klog.V(2).Infof("Creating IPPools")
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

func CreateOrUpdateKddCRD(c client.Client) error {
	for _, crd := range kddCrds {
		if err := utils.CreateOrUpdateEmbeddedCRD(c, crd); err != nil {
			klog.Errorf("Error creating the CRD: %v", err)
			return err
		}
	}
	return nil
}

func CreateOrUpdateServiceAccount(c client.Client) error {
	if err := serviceaccount.EnsureServiceAccount(c, "kube-system",
		embeddedyamls.Manifests_fix_calico_calico_kube_controllers_sa_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureServiceAccount(c, "kube-system",
		embeddedyamls.Manifests_fix_calico_calico_node_sa_yaml); err != nil {
		return err
	}
	return nil
}

func CreateOrUpdateClusterRole(c client.Client) error {
	if err := serviceaccount.EnsureClusterRole(c,
		embeddedyamls.Manifests_fix_calico_calico_kube_controllers_clusterrole_yaml); err != nil {
		return err
	}
	if err := serviceaccount.EnsureClusterRole(c,
		embeddedyamls.Manifests_fix_calico_calico_node_clusterrole_yaml); err != nil {
		return err
	}
	return nil
}

func CreateOrUpdateClusterRoleBinding(c client.Client) error {
	if err := serviceaccount.EnsureClusterRoleBinding(c, "kube-system",
		embeddedyamls.Manifests_fix_calico_calico_kube_controllers_clusterrolebinding_yaml); err != nil {
		return err
	}
	if err := serviceaccount.EnsureClusterRoleBinding(c, "kube-system",
		embeddedyamls.Manifests_fix_calico_calico_node_clusterrolebinding_yaml); err != nil {
		return err
	}
	return nil
}

func CreateOrUpdateDeployment(c client.Client) error {
	if err := deployments.EnsureDeployment(c, "kube-system",
		embeddedyamls.Manifests_fix_calico_calico_kube_controllers_yaml); err != nil {
		return err
	}
	return nil
}

func CreateOrUpdateDaemonSet(c client.Client) error {
	if err := daemonsets.EnsureDaemonSet(c, "kube-system",
		embeddedyamls.Manifests_fix_calico_calico_node_yaml); err != nil {
		return err
	}
	return nil
}

func CreateOrUpdateConfigMap(c client.Client) error {
	if err := configmaps.EnsureConfigMap(c, "kube-system",
		embeddedyamls.Manifests_fix_calico_calico_config_yaml); err != nil {
		return err
	}
	return nil
}

func CreateOrUpdatePodDisruptionBudget(c client.Client) error {
	if err := poddisruptionbudgets.EnsurePodDisruptionBudget(c, "kube-system",
		embeddedyamls.Manifests_fix_calico_calico_kube_controllers_pdb_yaml); err != nil {
		return err
	}
	return nil
}
