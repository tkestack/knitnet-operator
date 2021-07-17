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

package submarinerop

import (
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	consts "github.com/tkestack/knitnet-operator/controllers/ensures"
	"github.com/tkestack/knitnet-operator/controllers/ensures/common/namespace"
	lighthouseop "github.com/tkestack/knitnet-operator/controllers/ensures/operator/lighthouse"
	"github.com/tkestack/knitnet-operator/controllers/ensures/operator/submarinerop/crds"
	"github.com/tkestack/knitnet-operator/controllers/ensures/operator/submarinerop/deployment"
	"github.com/tkestack/knitnet-operator/controllers/ensures/operator/submarinerop/serviceaccount"
)

func Ensure(c client.Client, config *rest.Config, debug bool) error {
	if err := crds.Ensure(c); err != nil {
		return err
	}
	klog.Info("Created operator CRDs")

	if err := namespace.Ensure(c, consts.SubmarinerOperatorNamespace); err != nil {
		return err
	}

	if err := serviceaccount.Ensure(c, consts.SubmarinerOperatorNamespace); err != nil {
		return err
	}
	klog.Info("Created operator service account and role")

	if created, err := lighthouseop.Ensure(c, config, consts.SubmarinerOperatorNamespace); err != nil {
		return err
	} else if created {
		klog.Info("Created Lighthouse service accounts and roles")
	}

	if err := deployment.Ensure(c, consts.SubmarinerOperatorNamespace, consts.SubmarinerOperatorImage, debug); err != nil {
		return err
	}
	klog.Info("Deployed the operator successfully")
	return nil
}
