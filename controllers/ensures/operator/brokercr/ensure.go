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

package brokercr

import (
	"context"

	submariner "github.com/submariner-io/submariner-operator/apis/submariner/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	consts "github.com/tkestack/knitnet-operator/controllers/ensures"
)

func Ensure(c client.Client, brokerSpec submariner.BrokerSpec) error {
	brokerCR := &submariner.Broker{ObjectMeta: metav1.ObjectMeta{Name: consts.SubmarinerBrokerName, Namespace: consts.SubmarinerOperatorNamespace}}
	or, err := ctrl.CreateOrUpdate(context.TODO(), c, brokerCR, func() error {
		brokerCR.Spec = brokerSpec
		return nil
	})
	if err != nil {
		klog.Errorf("Failed to %s Broker %s: %v", or, brokerCR.GetName(), err)
		return err
	}
	klog.Infof("Broker %s %s", brokerCR.GetName(), or)
	return nil
}
