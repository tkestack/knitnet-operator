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
package broker

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	GlobalCIDRConfigMapName = "submariner-globalnet-info"
	GlobalnetStatusKey      = "globalnetEnabled"
	ClusterInfoKey          = "clusterinfo"
	GlobalnetCidrRange      = "globalnetCidrRange"
	GlobalnetClusterSize    = "globalnetClusterSize"
)

type ClusterInfo struct {
	ClusterID     string   `json:"cluster_id"`
	NetworkPlugin string   `json:"network_plugin"`
	GlobalCidr    []string `json:"global_cidr"`
}

func CreateGlobalnetConfigMap(c client.Client, globalnetEnabled bool, defaultGlobalCidrRange string,
	defaultGlobalClusterSize uint, namespace string) error {
	klog.Info("Create or update globalnet configmap")
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GlobalCIDRConfigMapName,
			Namespace: namespace,
		},
	}
	or, err := ctrl.CreateOrUpdate(context.TODO(), c, cm, func() error {
		return GeneralGlobalnetConfigMap(cm, globalnetEnabled, defaultGlobalCidrRange, defaultGlobalClusterSize)
	})
	if err != nil {
		klog.Errorf("error %s globalnet configmap: %v", or, err)
	}
	klog.Infof("Configmap %s %s", GlobalCIDRConfigMapName, or)
	return nil
}

func GeneralGlobalnetConfigMap(cm *v1.ConfigMap, globalnetEnabled bool, defaultGlobalCidrRange string, defaultGlobalClusterSize uint) error {
	labels := map[string]string{
		"component": "submariner-globalnet",
	}
	cidrRange, err := json.Marshal(defaultGlobalCidrRange)
	if err != nil {
		return err
	}

	var data map[string]string
	if globalnetEnabled {
		data = map[string]string{
			GlobalnetStatusKey:   "true",
			GlobalnetCidrRange:   string(cidrRange),
			GlobalnetClusterSize: fmt.Sprint(defaultGlobalClusterSize),
			ClusterInfoKey:       "[]",
		}
	} else {
		data = map[string]string{
			GlobalnetStatusKey: "false",
			ClusterInfoKey:     "[]",
		}
	}
	cm.ObjectMeta.Labels = labels
	cm.Data = data
	return nil
}

func UpdateGlobalnetConfigMap(c client.Client, namespace string,
	configMap *v1.ConfigMap, newCluster ClusterInfo) error {
	var clusterInfos []ClusterInfo
	err := json.Unmarshal([]byte(configMap.Data[ClusterInfoKey]), &clusterInfos)
	if err != nil {
		return err
	}

	exists := false
	for k, value := range clusterInfos {
		if value.ClusterID == newCluster.ClusterID {
			clusterInfos[k].GlobalCidr = newCluster.GlobalCidr
			exists = true
		}
	}

	if !exists {
		// var newEntry ClusterInfo
		// newEntry.ClusterID = newCluster.ClusterID
		// newEntry.NetworkPlugin = newCluster.NetworkPlugin
		// newEntry.GlobalCidr = newCluster.GlobalCidr
		clusterInfos = append(clusterInfos, newCluster)
	}

	data, err := json.MarshalIndent(clusterInfos, "", "\t")
	if err != nil {
		return err
	}

	configMap.Data[ClusterInfoKey] = string(data)
	return c.Update(context.TODO(), configMap)
}

func GetGlobalnetConfigMap(reader client.Reader, namespace string) (*v1.ConfigMap, error) {
	cm := &v1.ConfigMap{}
	cmKey := types.NamespacedName{Name: GlobalCIDRConfigMapName, Namespace: namespace}
	if err := reader.Get(context.TODO(), cmKey, cm); err != nil {
		return nil, err
	}
	return cm, nil
}

func GetClusterInfos(reader client.Reader, namespace string) ([]ClusterInfo, error) {
	cm, err := GetGlobalnetConfigMap(reader, namespace)
	if err != nil {
		return nil, err
	}
	var clusterInfos []ClusterInfo
	err = json.Unmarshal([]byte(cm.Data[ClusterInfoKey]), &clusterInfos)
	if err != nil {
		return nil, err
	}
	return clusterInfos, nil
}
