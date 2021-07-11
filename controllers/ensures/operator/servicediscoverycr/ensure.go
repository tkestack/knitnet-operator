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

package servicediscoverycr

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	submariner "github.com/submariner-io/submariner-operator/apis/submariner/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/tkestack/knitnet-operator/controllers/ensures/names"
)

func Ensure(c client.Client, namespace string, serviceDiscoverySpec *submariner.ServiceDiscoverySpec) error {
	sd := &submariner.ServiceDiscovery{ObjectMeta: metav1.ObjectMeta{Name: names.ServiceDiscoveryCrName, Namespace: namespace}}
	or, err := ctrl.CreateOrUpdate(context.TODO(), c, sd, func() error {
		sd.Spec = *serviceDiscoverySpec
		return nil
	})
	if err != nil {
		klog.Errorf("Failed to %s ServiceDiscovery %s: %v", or, sd.GetName(), err)
		return err
	}
	klog.Infof("ServiceDiscovery %s %s", sd.GetName(), or)
	return nil
}
