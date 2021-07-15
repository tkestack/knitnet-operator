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

package controllers

import (
	submarinerv1a1 "github.com/submariner-io/submariner-operator/apis/submariner/v1alpha1"
	"k8s.io/klog/v2"

	"github.com/tkestack/knitnet-operator/controllers/components"
	consts "github.com/tkestack/knitnet-operator/controllers/ensures"

	"github.com/tkestack/knitnet-operator/controllers/discovery/globalnet"
	"github.com/tkestack/knitnet-operator/controllers/ensures/broker"
	"github.com/tkestack/knitnet-operator/controllers/ensures/operator/brokercr"

	operatorv1alpha1 "github.com/tkestack/knitnet-operator/api/v1alpha1"
	"github.com/tkestack/knitnet-operator/controllers/ensures/operator/submarinerop"
)

// var defaultComponents = []string{components.ServiceDiscovery, components.Connectivity}
// var validComponents = []string{components.ServiceDiscovery, components.Connectivity, components.Globalnet, components.Broker}

func (r *KnitnetReconciler) DeploySubmerinerBroker(instance *operatorv1alpha1.Knitnet) error {
	brokerConfig := &instance.Spec.BrokerConfig

	// if err := isValidComponents(instance); err != nil {
	// 	klog.Errorf("Invalid components parameter: %v", err)
	// 	return err
	// }

	if valid, err := isValidGlobalnetConfig(instance); !valid {
		klog.Errorf("Invalid GlobalCIDR configuration: %v", err)
		return err
	}

	klog.Info("Setting up broker RBAC")
	if err := broker.Ensure(r.Client, r.Config, brokerConfig.ServiceDiscoveryEnabled, brokerConfig.GlobalnetEnable, false); err != nil {
		klog.Errorf("Error setting up broker RBAC: %v", err)
		return err
	}
	klog.Info("Deploying the Submariner operator")
	if err := submarinerop.Ensure(r.Client, r.Config, true); err != nil {
		klog.Errorf("Error deploying the operator: %v", err)
		return err
	}
	klog.Info("Deploying the broker")
	if err := brokercr.Ensure(r.Client, populateBrokerSpec(instance)); err != nil {
		klog.Errorf("Broker deployment failed: %v", err)
		return err
	}

	if err := broker.CreateGlobalnetConfigMap(r.Client, brokerConfig.GlobalnetEnable, brokerConfig.GlobalnetCIDRRange,
		brokerConfig.DefaultGlobalnetClusterSize, consts.SubmarinerBrokerNamespace); err != nil {
		klog.Errorf("Error creating globalCIDR configmap on Broker: %v", err)
		return err
	}

	if brokerConfig.GlobalnetEnable {
		if err := globalnet.ValidateExistingGlobalNetworks(r.Reader, consts.SubmarinerBrokerNamespace); err != nil {
			klog.Errorf("Error validating existing globalCIDR configmap: %v", err)
			return err
		}
	}

	if err := broker.CreateBrokerInfoConfigMap(r.Client, r.Config, instance); err != nil {
		klog.Errorf("Error writing the broker information: %v", err)
		return err
	}
	return nil
}

// func isValidComponents(instance *operatorv1alpha1.Knitnet) error {
// 	componentSet := stringset.New(instance.Spec.BrokerConfig.ComponentArr...)
// 	validComponentSet := stringset.New(validComponents...)

// 	if componentSet.Size() < 1 {
// 		klog.Info("Use default components")
// 		instance.Spec.BrokerConfig.ComponentArr = defaultComponents
// 		return nil
// 	}

// 	for _, component := range componentSet.Elements() {
// 		if !validComponentSet.Contains(component) {
// 			return fmt.Errorf("unknown component: %s", component)
// 		}
// 	}
// 	return nil
// }

func isValidGlobalnetConfig(instance *operatorv1alpha1.Knitnet) (bool, error) {
	brokerConfig := &instance.Spec.BrokerConfig
	var err error
	if !brokerConfig.GlobalnetEnable {
		return true, nil
	}
	defaultGlobalnetClusterSize, err := globalnet.GetValidClusterSize(brokerConfig.GlobalnetCIDRRange, brokerConfig.DefaultGlobalnetClusterSize)
	if err != nil || defaultGlobalnetClusterSize == 0 {
		return false, err
	}
	return true, err
}

func populateBrokerSpec(instance *operatorv1alpha1.Knitnet) submarinerv1a1.BrokerSpec {
	brokerConfig := instance.Spec.BrokerConfig
	enabledComponents := []string{}
	if brokerConfig.ConnectivityEnabled {
		enabledComponents = append(enabledComponents, components.Connectivity)
	}
	if brokerConfig.GlobalnetEnable {
		enabledComponents = append(enabledComponents, components.Globalnet)
	}
	if brokerConfig.ServiceDiscoveryEnabled {
		enabledComponents = append(enabledComponents, components.ServiceDiscovery)
	}
	brokerSpec := submarinerv1a1.BrokerSpec{
		GlobalnetEnabled:            brokerConfig.GlobalnetEnable,
		GlobalnetCIDRRange:          brokerConfig.GlobalnetCIDRRange,
		DefaultGlobalnetClusterSize: brokerConfig.DefaultGlobalnetClusterSize,
		Components:                  enabledComponents,
		DefaultCustomDomains:        brokerConfig.DefaultCustomDomains,
	}
	return brokerSpec
}
