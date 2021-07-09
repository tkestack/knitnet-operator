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

package broker

import (
	"context"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	consts "github.com/tkestack/cluster-fabric-operator/controllers/ensures"
)

const (
	submarinerBrokerClusterRole      = "submariner-k8s-broker-cluster"
	submarinerBrokerAdminRole        = "submariner-k8s-broker-admin"
	SubmarinerBrokerAdminSA          = "submariner-k8s-broker-admin"
	submarinerBrokerClusterSAFmt     = "cluster-%s"
	submarinerBrokerClusterDefaultSA = "submariner-k8s-broker-client" // for backwards compatibility with documentation
)

func NewBrokerSA(submarinerBrokerSA string) *v1.ServiceAccount {
	sa := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      submarinerBrokerSA,
			Namespace: consts.SubmarinerBrokerNamespace,
		},
	}

	return sa
}

// Create a role for to bind the cluster admin SA
func NewBrokerRoleBinding(serviceAccount, role string) *rbacv1.RoleBinding {
	binding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", serviceAccount, role),
			Namespace: consts.SubmarinerBrokerNamespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     role,
		},
		Subjects: []rbacv1.Subject{
			{
				Namespace: "submariner-k8s-broker",
				Name:      serviceAccount,
				Kind:      "ServiceAccount",
			},
		},
	}

	return binding
}

func GetClientTokenSecret(reader client.Reader, brokerNamespace, submarinerBrokerSA string) (*v1.Secret, error) {
	sa := &v1.ServiceAccount{}
	saKey := types.NamespacedName{Name: submarinerBrokerSA, Namespace: brokerNamespace}
	if err := reader.Get(context.TODO(), saKey, sa); err != nil {
		klog.Errorf("ServiceAccount %s get failed: %v", submarinerBrokerSA, err)
		return nil, err
	}
	if len(sa.Secrets) < 1 {
		klog.Errorf("ServiceAccount %s does not have any secret", sa.Name)
		return nil, fmt.Errorf("ServiceAccount %s does not have any secret", sa.Name)
	}
	brokerTokenPrefix := fmt.Sprintf("%s-token-", submarinerBrokerSA)

	for _, secret := range sa.Secrets {
		if strings.HasPrefix(secret.Name, brokerTokenPrefix) {
			sec := &v1.Secret{}
			secKey := types.NamespacedName{Name: secret.Name, Namespace: brokerNamespace}
			err := reader.Get(context.TODO(), secKey, sec)
			return sec, err
		}
	}

	return nil, fmt.Errorf("ServiceAccount %s does not have a secret of type token", submarinerBrokerSA)
}
