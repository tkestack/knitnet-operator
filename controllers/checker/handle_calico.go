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
	"bytes"
	"context"
	"text/template"

	submarinerv1 "github.com/submariner-io/submariner/pkg/apis/submariner.io/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	netconsts "github.com/tkestack/knitnet-operator/controllers/discovery"
	"github.com/tkestack/knitnet-operator/controllers/ensures/broker"
)

func EnsureCalico(c client.Client, currentClusterID string, clusterInfos *[]broker.ClusterInfo) error {
	klog.Infof("Creating IPPools")
	clusters, err := getClusters(c)
	if err != nil {
		return err
	}
	for _, clusterInfo := range *clusterInfos {
		if clusterInfo.ClusterID == currentClusterID || clusterInfo.NetworkPlugin != netconsts.NetworkPluginCalico {
			continue
		}
		cluster := getClusterWithID(clusterInfo.ClusterID, clusters)
		if err := createOrUpdateIPPools(c, "pod-cidr-"+cluster.Spec.ClusterID, cluster.Spec.ClusterCIDR[0]); err != nil {
			return err
		}
		if err := createOrUpdateIPPools(c, "svc-cidr-"+cluster.Spec.ClusterID, cluster.Spec.ServiceCIDR[0]); err != nil {
			return err
		}
	}
	return nil
}

func getClusterWithID(ID string, clusters *submarinerv1.ClusterList) *submarinerv1.Cluster {
	for _, cluster := range clusters.Items {
		if cluster.Spec.ClusterID == ID {
			return &cluster
		}
	}
	return nil
}

func getClusters(c client.Client) (*submarinerv1.ClusterList, error) {
	clusters := &submarinerv1.ClusterList{}
	if err := c.List(context.TODO(), clusters); err != nil {
		klog.Errorf("Failed to list Cluster: %v", err)
		return nil, err
	}
	return clusters, nil
}

const ippool = `
---
apiVersion: crd.projectcalico.org/v1
kind: IPPool
metadata:
  name: {{ .NAME }}
spec:
  cidr: {{ .CIDR }}
  natOutgoing: false
  disabled: true
`

type IPPoolData struct {
	NAME string
	CIDR string
}

func createOrUpdateIPPools(c client.Client, name, cidr string) error {
	ippoolData := IPPoolData{
		NAME: name,
		CIDR: cidr,
	}
	var ippoolYaml bytes.Buffer
	t := template.Must(template.New("ippool").Parse(ippool))
	if err := t.Execute(&ippoolYaml, ippoolData); err != nil {
		return err
	}
	klog.Infof("Create or update IPPool %s", name)
	if err := createUpdateFromYaml(c, ippoolYaml.Bytes()); err != nil {
		return err
	}
	return nil
}

func createUpdateFromYaml(c client.Client, yamlContent []byte) error {
	obj := &unstructured.Unstructured{}
	jsonSpec, err := yaml.YAMLToJSON(yamlContent)
	if err != nil {
		klog.Errorf("could not convert yaml to json: %v", err)
		return err
	}

	if err := obj.UnmarshalJSON(jsonSpec); err != nil {
		klog.Errorf("could not unmarshal resource: %v", err)
		return err
	}

	or, err := ctrl.CreateOrUpdate(context.TODO(), c, obj, func() error {
		return nil
	})
	if err != nil {
		klog.Errorf("Failed to %s Object %s: %v", or, obj.GetName(), err)
		return err
	}
	klog.Infof("Object %s %s", obj.GetName(), or)
	return nil
}
