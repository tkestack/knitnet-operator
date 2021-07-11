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
package lighthouse

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tkestack/knitnet-operator/controllers/ensures/operator/common/embeddedyamls"
	"github.com/tkestack/knitnet-operator/controllers/ensures/utils"
)

const (
	BrokerCluster = true
	DataCluster   = false
)

// Ensure ensures that the required resources are deployed on the target system
// The resources handled here are the lighthouse CRDs: MultiClusterService,
// ServiceImport, ServiceExport and ServiceDiscovery
func Ensure(crdUpdater utils.CRDUpdater, c client.Client, isBroker bool) error {
	// Delete obsolete CRDs if they are still present

	err := crdUpdater.Delete(context.TODO(), "serviceimports.lighthouse.submariner.io", metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("error deleting the obsolete ServiceImport CRD: %v", err)
		return err
	}
	err = crdUpdater.Delete(context.TODO(), "serviceexports.lighthouse.submariner.io", metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("error deleting the obsolete ServiceExport CRD: %v", err)
		return err
	}
	err = crdUpdater.Delete(context.TODO(), "multiclusterservices.lighthouse.submariner.io", metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("error deleting the obsolete MultiClusterServices CRD: %v", err)
		return err
	}

	if err := utils.CreateOrUpdateEmbeddedCRD(c,
		embeddedyamls.Manifests_deploy_mcsapi_crds_multicluster_x_k8s_io_serviceimports_yaml); err != nil {
		klog.Errorf("error creating the MCS ServiceImport CRD: %v", err)
		return err
	}

	// The broker does not need the ServiceExport or ServiceDiscovery
	if isBroker {
		return nil
	}

	if err := utils.CreateOrUpdateEmbeddedCRD(c,
		embeddedyamls.Manifests_deploy_mcsapi_crds_multicluster_x_k8s_io_serviceexports_yaml); err != nil {
		klog.Errorf("error creating the MCS ServiceExport CRD: %v", err)
		return err
	}

	if err := utils.CreateOrUpdateEmbeddedCRD(c, embeddedyamls.Manifests_deploy_crds_submariner_io_servicediscoveries_yaml); err != nil {
		klog.Errorf("error creating the ServiceDiscovery CRD: %v", err)
		return err
	}

	return nil
}
