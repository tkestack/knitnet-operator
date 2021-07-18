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

package daemonsets

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tkestack/knitnet-operator/controllers/embeddedyamls"
)

func EnsureDaemonSet(c client.Client, namespace, yaml string) error {
	dsName, err := embeddedyamls.GetObjectName(yaml)
	if err != nil {
		return err
	}
	ds := &appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Name: dsName, Namespace: namespace}}
	or, err := ctrl.CreateOrUpdate(context.TODO(), c, ds, func() error {
		if err := embeddedyamls.GetObject(yaml, ds); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		klog.Errorf("Failed to %s DaemonSet %s: %v", or, ds.GetName(), err)
		return err
	}
	klog.V(2).Infof("DaemonSet %s %s", ds.GetName(), or)
	return nil
}
