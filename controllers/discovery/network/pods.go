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

package network

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func findPodCommandParameter(c client.Client, labelSelector, parameter string) (string, error) {
	pod, err := findPod(c, labelSelector)

	if err != nil || pod == nil {
		return "", err
	}
	for _, container := range pod.Spec.Containers {
		for _, arg := range container.Command {
			if strings.HasPrefix(arg, parameter) {
				return strings.Split(arg, "=")[1], nil
			}
			// Handling the case where the command is in the form of /bin/sh -c exec ....
			if strings.Contains(arg, " ") {
				for _, subArg := range strings.Split(arg, " ") {
					if strings.HasPrefix(subArg, parameter) {
						return strings.Split(subArg, "=")[1], nil
					}
				}
			}
		}
	}
	return "", nil
}

func findPod(c client.Client, labelSelector string) (*v1.Pod, error) {
	pods := &v1.PodList{}
	// 	matchingLabels := client.MatchingLabels(map[string]string{"k": "axahm2EJ8Phiephe2eixohbee9eGeiyees1thuozi1xoh0GiuH3diewi8iem7Nui"})
	// listOpts := &client.ListOptions{}
	// matchingLabels.ApplyToList(listOpts)
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		return nil, err
	}
	opts := &client.ListOptions{
		LabelSelector: selector,
		Limit:         1,
	}
	if err := c.List(context.TODO(), pods, opts); err != nil {
		return nil, errors.WithMessagef(err, "error listing Pods by label selector %q", labelSelector)
	}

	if len(pods.Items) == 0 {
		return nil, nil
	}

	return &pods.Items[0], nil
}
