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

package submarinercr

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	submariner "github.com/submariner-io/submariner-operator/apis/submariner/v1alpha1"
)

const (
	SubmarinerName = "submariner"
)

var backOff wait.Backoff = wait.Backoff{
	Steps:    20,
	Duration: 1 * time.Second,
	Factor:   1.5,
	Cap:      60 * time.Second,
}

func Ensure(c client.Client, namespace string, submarinerSpec *submariner.SubmarinerSpec) error {
	newSubmarinerCR := &submariner.Submariner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SubmarinerName,
			Namespace: namespace,
		},
		Spec: *submarinerSpec,
	}
	submarinerCR := &submariner.Submariner{}
	submarinerCRKey := types.NamespacedName{Name: SubmarinerName, Namespace: namespace}
	return wait.ExponentialBackoff(backOff, func() (bool, error) {
		if err := c.Get(context.TODO(), submarinerCRKey, submarinerCR); err != nil {
			if errors.IsNotFound(err) {
				klog.Info("Creating new submerinerCR")
				if err := c.Create(context.TODO(), newSubmarinerCR); err != nil {
					return false, err
				}
				return true, nil
			}
			return false, err
		}

		if !submarinerCR.ObjectMeta.DeletionTimestamp.IsZero() {
			klog.Info("SubmerinerCR is deleted, waiting for the delete complete...")
			return false, nil
		}

		klog.Info("Try to delete existing submerinerCR")
		fg := metav1.DeletePropagationForeground
		delOpts := &client.DeleteOptions{PropagationPolicy: &fg}
		err := c.Delete(context.TODO(), submarinerCR, delOpts)
		return false, client.IgnoreNotFound(err)
	})
}
