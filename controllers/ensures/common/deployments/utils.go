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

package deployments

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WaitForReady(c client.Client, namespace, deployment string, interval, timeout time.Duration) error {
	// deployments := clientSet.AppsV1().Deployments(namespace)
	deploy := &appsv1.Deployment{}
	deployKey := types.NamespacedName{Name: deployment, Namespace: namespace}
	if err := c.Get(context.TODO(), deployKey, deploy); err != nil {
		return err
	}
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		err := c.Get(context.TODO(), deployKey, deploy)
		if err != nil && !errors.IsNotFound(err) {
			return false, fmt.Errorf("error waiting for controller deployment to come up: %s", err)
		}

		for _, cond := range deploy.Status.Conditions {
			if cond.Type == appsv1.DeploymentAvailable && cond.Status == v1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
}
