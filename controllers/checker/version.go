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

package checker

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

const (
	minK8sMajor  = 1  // We need K8s 1.15 for endpoint slices
	minK8sMinor  = 15 // Need patch if k8s minor version < 17 and >= 15
	goodK8sMinor = 17 // Don't need patch if k8s minor version >= 17
)

func CheckKubernetesVersion(config *rest.Config) (bool, error) {
	needPatch := false
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorf("Error creating API server client: %v", err)
		return needPatch, err
	}
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		klog.Errorf("Error obtaining API server version: %v", err)
		return needPatch, err
	}
	major, err := strconv.Atoi(serverVersion.Major)
	if err != nil {
		klog.Errorf("Error parsing API server major version %v", err)
		return needPatch, err
	}
	var minor int
	if strings.HasSuffix(serverVersion.Minor, "+") {
		minor, err = strconv.Atoi(serverVersion.Minor[0 : len(serverVersion.Minor)-1])
	} else {
		minor, err = strconv.Atoi(serverVersion.Minor)
	}
	if err != nil {
		klog.Errorf("Error parsing API server minor version %v", err)
		return needPatch, err
	}

	if major < minK8sMajor || (major == minK8sMajor && minor < minK8sMinor) {
		klog.Errorf("Submariner requires Kubernetes %d.%d; your cluster is running %s.%s",
			minK8sMajor, minK8sMinor, serverVersion.Major, serverVersion.Minor)
		return needPatch, fmt.Errorf("submariner requires Kubernetes %d.%d; your cluster is running %s.%s",
			minK8sMajor, minK8sMinor, serverVersion.Major, serverVersion.Minor)
	} else {
		if minor < goodK8sMinor {
			needPatch = true
		}
		return needPatch, nil
	}
}
