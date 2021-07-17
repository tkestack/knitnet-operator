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

package ippools

import (
	"context"

	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	"github.com/projectcalico/api/pkg/client/clientset_generated/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

func EnsureIPPool(config *rest.Config, name, cidr string) error {
	cs, err := clientset.NewForConfig(config)
	if err != nil {
		return err
	}
	ipPool := &v3.IPPool{ObjectMeta: metav1.ObjectMeta{Name: name}}
	ipPool.Spec.CIDR = cidr
	_, err = cs.ProjectcalicoV3().IPPools().Create(context.TODO(), &v3.IPPool{ObjectMeta: metav1.ObjectMeta{Name: name}}, metav1.CreateOptions{})
	klog.Errorf("Create IPPool %s failed: %v", name, err)
	return err
}
