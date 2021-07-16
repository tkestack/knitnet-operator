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

package crds

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tkestack/knitnet-operator/controllers/embeddedyamls"
	"github.com/tkestack/knitnet-operator/controllers/utils"
)

// Ensure functions updates or installs the operator CRDs in the cluster
func Ensure(c client.Client) error {
	// Attempt to update or create the CRD definitions
	// TODO(majopela): In the future we may want to report when we have updated the existing
	//                 CRD definition with new versions
	if err := utils.CreateOrUpdateEmbeddedCRD(c, embeddedyamls.Manifests_deploy_crds_submariner_io_submariners_yaml); err != nil {
		return err
	}
	if err := utils.CreateOrUpdateEmbeddedCRD(c,
		embeddedyamls.Manifests_deploy_crds_submariner_io_servicediscoveries_yaml); err != nil {
		return err
	}
	if err := utils.CreateOrUpdateEmbeddedCRD(c, embeddedyamls.Manifests_deploy_crds_submariner_io_brokers_yaml); err != nil {
		return err
	}
	return nil
}
