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

package configmaps

import (
	"context"

	"github.com/tkestack/knitnet-operator/controllers/embeddedyamls"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func EnsureConfigMap(c client.Client, namespace, yaml string) error {
	cmName, err := embeddedyamls.GetObjectName(yaml)
	if err != nil {
		return err
	}
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: cmName, Namespace: namespace}}
	or, err := ctrl.CreateOrUpdate(context.TODO(), c, cm, func() error {
		if err := embeddedyamls.GetObject(yaml, cm); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		klog.Errorf("Failed to %s ConfigMap %s: %v", or, cm.GetName(), err)
		return err
	}
	klog.V(2).Infof("ConfigMap %s %s", cm.GetName(), or)
	return nil
}
