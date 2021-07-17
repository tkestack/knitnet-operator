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

package namespace

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Ensure functions updates or installs the operator CRDs in the cluster
func Ensure(c client.Client, name string) error {
	ns := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}

	or, err := ctrl.CreateOrUpdate(context.TODO(), c, ns, func() error {
		return nil
	})
	if err != nil {
		klog.Errorf("Namespace %s %s failed: %v", name, or, err)
	}
	klog.Infof("Namespace %s %s", name, or)
	return nil
}
