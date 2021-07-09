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

package lighthouseop

import (
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tkestack/cluster-fabric-operator/controllers/ensures/operator/lighthouse/scc"
	"github.com/tkestack/cluster-fabric-operator/controllers/ensures/operator/lighthouse/serviceaccount"
)

func Ensure(c client.Client, config *rest.Config, operatorNamespace string) (bool, error) {
	if err := serviceaccount.Ensure(c, operatorNamespace); err != nil {
		return false, err
	}

	if created, err := scc.Ensure(config, operatorNamespace); err != nil {
		return created, err
	} else if created {
		klog.Info("Updated the privileged SCC")
	}

	return true, nil
}
