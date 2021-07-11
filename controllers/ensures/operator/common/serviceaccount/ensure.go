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

package serviceaccount

import (
	"context"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tkestack/knitnet-operator/controllers/ensures/operator/common/embeddedyamls"
)

// Ensure creates the given service account
func Ensure(c client.Client, namespace, yaml string) error {
	saName, err := embeddedyamls.GetObjectName(yaml)
	if err != nil {
		return err
	}
	sa := &v1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: saName, Namespace: namespace}}
	or, err := ctrl.CreateOrUpdate(context.TODO(), c, sa, func() error {
		if err := embeddedyamls.GetObject(yaml, sa); err != nil {
			return err
		}
		// sa.SetNamespace(namespace)
		return nil
	})
	if err != nil {
		klog.Errorf("Failed to %s ServiceAccount %s: %v", or, sa.GetName(), err)
		return err
	}
	klog.V(2).Infof("ServiceAccount %s %s", sa.GetName(), or)
	return nil
}

func EnsureRole(c client.Client, namespace, yaml string) error {
	roleName, err := embeddedyamls.GetObjectName(yaml)
	if err != nil {
		return err
	}
	role := &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: roleName, Namespace: namespace}}
	or, err := ctrl.CreateOrUpdate(context.TODO(), c, role, func() error {
		if err := embeddedyamls.GetObject(yaml, role); err != nil {
			return err
		}
		// role.SetNamespace(namespace)
		return nil
	})
	if err != nil {
		klog.Errorf("Failed to %s role %s: %v", or, role.GetName(), err)
		return err
	}
	klog.V(2).Infof("Role %s %s", role.GetName(), or)
	return nil
}

func EnsureRoleBinding(c client.Client, namespace, yaml string) error {
	roleBindingName, err := embeddedyamls.GetObjectName(yaml)
	if err != nil {
		return err
	}
	roleBinding := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: roleBindingName, Namespace: namespace}}
	or, err := ctrl.CreateOrUpdate(context.TODO(), c, roleBinding, func() error {
		if err := embeddedyamls.GetObject(yaml, roleBinding); err != nil {
			return err
		}
		// roleBinding.SetNamespace(namespace)
		return nil
	})
	if err != nil {
		klog.Errorf("Failed to %s RoleBinding %s: %v", or, roleBinding.GetName(), err)
		return err
	}
	klog.V(2).Infof("RoleBinding %s %s", roleBinding.GetName(), or)
	return nil
}

func EnsureClusterRole(c client.Client, yaml string) error {
	clusterRoleName, err := embeddedyamls.GetObjectName(yaml)
	if err != nil {
		return err
	}
	clusterRole := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: clusterRoleName}}

	or, err := ctrl.CreateOrUpdate(context.TODO(), c, clusterRole, func() error {
		return embeddedyamls.GetObject(yaml, clusterRole)
	})
	if err != nil {
		klog.Errorf("Failed to %s ClusterRole %s: %v", or, clusterRole.GetName(), err)
		return err
	}
	klog.V(2).Infof("ClusterRole %s %s", clusterRole.GetName(), or)
	return nil
}

func EnsureClusterRoleBinding(c client.Client, namespace, yaml string) error {
	clusterRoleBindingName, err := embeddedyamls.GetObjectName(yaml)
	if err != nil {
		return err
	}
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: clusterRoleBindingName}}
	or, err := ctrl.CreateOrUpdate(context.TODO(), c, clusterRoleBinding, func() error {
		if err := embeddedyamls.GetObject(yaml, clusterRoleBinding); err != nil {
			return err
		}
		clusterRoleBinding.Subjects[0].Namespace = namespace
		return nil
	})
	if err != nil {
		klog.Errorf("Failed to %s ClusterRoleBinding %s: %v", or, clusterRoleBinding.GetName(), err)
		return err
	}
	klog.V(2).Infof("ClusterRoleBinding %s %s", clusterRoleBinding.GetName(), or)
	return nil
}
