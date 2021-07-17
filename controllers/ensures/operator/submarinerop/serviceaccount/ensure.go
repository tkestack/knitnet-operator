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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tkestack/knitnet-operator/controllers/ensures/common/serviceaccount"

	"github.com/tkestack/knitnet-operator/controllers/embeddedyamls"
)

// Ensure functions updates or installs the operator CRDs in the cluster
func Ensure(c client.Client, namespace string) error {
	if err := ensureServiceAccounts(c, namespace); err != nil {
		return err
	}

	if err := ensureRoles(c, namespace); err != nil {
		return err
	}

	if err := ensureRoleBindings(c, namespace); err != nil {
		return err
	}

	if err := ensureClusterRoles(c); err != nil {
		return err
	}

	if err := ensureClusterRoleBindings(c, namespace); err != nil {
		return err
	}

	return nil
}

func ensureServiceAccounts(c client.Client, namespace string) error {
	if err := serviceaccount.EnsureServiceAccount(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_operator_service_account_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureServiceAccount(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_gateway_service_account_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureServiceAccount(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_route_agent_service_account_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureServiceAccount(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_globalnet_service_account_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureServiceAccount(c, namespace,
		embeddedyamls.Manifests_config_rbac_networkplugin_syncer_service_account_yaml); err != nil {
		return err
	}
	return nil
}

func ensureClusterRoles(c client.Client) error {
	if err := serviceaccount.EnsureClusterRole(c,
		embeddedyamls.Manifests_config_rbac_submariner_operator_cluster_role_yaml); err != nil {
		return err
	}
	if err := serviceaccount.EnsureClusterRole(c,
		embeddedyamls.Manifests_config_rbac_submariner_gateway_cluster_role_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureClusterRole(c,
		embeddedyamls.Manifests_config_rbac_submariner_route_agent_cluster_role_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureClusterRole(c,
		embeddedyamls.Manifests_config_rbac_submariner_globalnet_cluster_role_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureClusterRole(c,
		embeddedyamls.Manifests_config_rbac_networkplugin_syncer_cluster_role_yaml); err != nil {
		return err
	}
	return nil
}

func ensureClusterRoleBindings(c client.Client, namespace string) error {
	if err := serviceaccount.EnsureClusterRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_operator_cluster_role_binding_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureClusterRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_gateway_cluster_role_binding_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureClusterRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_route_agent_cluster_role_binding_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureClusterRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_globalnet_cluster_role_binding_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureClusterRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_networkplugin_syncer_cluster_role_binding_yaml); err != nil {
		return err
	}

	return nil
}

func ensureRoles(c client.Client, namespace string) error {
	if err := serviceaccount.EnsureRole(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_operator_role_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureRole(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_gateway_role_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureRole(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_route_agent_role_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureRole(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_globalnet_role_yaml); err != nil {
		return err
	}

	return nil
}

func ensureRoleBindings(c client.Client, namespace string) error {
	if err := serviceaccount.EnsureRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_operator_role_binding_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_gateway_role_binding_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_route_agent_role_binding_yaml); err != nil {
		return err
	}

	if err := serviceaccount.EnsureRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_globalnet_role_binding_yaml); err != nil {
		return err
	}

	return nil
}
