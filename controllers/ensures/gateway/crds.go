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

package gateway

import (
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tkestack/knitnet-operator/controllers/ensures/operator/common/embeddedyamls"
	"github.com/tkestack/knitnet-operator/controllers/ensures/utils"
)

// Ensure ensures that the required resources are deployed on the target system
// The resources handled here are the gateway CRDs: Cluster and Endpoint
func Ensure(c client.Client) error {
	if err := utils.CreateOrUpdateEmbeddedCRD(c, embeddedyamls.Manifests_deploy_submariner_crds_submariner_io_clusters_yaml); err != nil {
		klog.Errorf("error provisioning the Cluster CRD: %v", err)
		return err
	}
	if err := utils.CreateOrUpdateEmbeddedCRD(c, embeddedyamls.Manifests_deploy_submariner_crds_submariner_io_endpoints_yaml); err != nil {
		klog.Errorf("error provisioning the Endpoint CRD: %v", err)
		return err
	}
	if err := utils.CreateOrUpdateEmbeddedCRD(c, embeddedyamls.Manifests_deploy_submariner_crds_submariner_io_gateways_yaml); err != nil {
		klog.Errorf("error provisioning the Gateway CRD: %v", err)
		return err
	}
	return nil
}
