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
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	submariner "github.com/submariner-io/submariner-operator/apis/submariner/v1alpha1"

	operatorv1alpha1 "github.com/tkestack/knitnet-operator/api/v1alpha1"
	cmdVersion "github.com/tkestack/knitnet-operator/controllers/checker"
	"github.com/tkestack/knitnet-operator/controllers/discovery/globalnet"
	"github.com/tkestack/knitnet-operator/controllers/discovery/network"
	consts "github.com/tkestack/knitnet-operator/controllers/ensures"
	"github.com/tkestack/knitnet-operator/controllers/ensures/broker"
	"github.com/tkestack/knitnet-operator/controllers/ensures/names"
	"github.com/tkestack/knitnet-operator/controllers/ensures/operator/servicediscoverycr"
	"github.com/tkestack/knitnet-operator/controllers/ensures/operator/submarinercr"
	"github.com/tkestack/knitnet-operator/controllers/ensures/operator/submarinerop"
	"github.com/tkestack/knitnet-operator/controllers/versions"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var clienttoken *v1.Secret

var nodeLabelBackoff wait.Backoff = wait.Backoff{
	Steps:    10,
	Duration: 1 * time.Second,
	Factor:   1.2,
	Jitter:   1,
}

func (r *KnitnetReconciler) JoinSubmarinerCluster(instance *operatorv1alpha1.Knitnet, brokerInfo *broker.BrokerInfo) error {
	joinConfig := instance.Spec.JoinConfig

	if err := isValidCustomCoreDNSConfig(instance); err != nil {
		klog.Errorf("Invalid Custom CoreDNS configuration: %v", err)
		return err
	}

	if joinConfig.ClusterID == "" {
		// r.Config
		// rawConfig, err := r.Config.RawConfig()
		// // This will be fatal later, no point in continuing
		// utils.ExitOnError("Error connecting to the target cluster", err)
		// clusterName := restconfig.ClusterNameFromContext(rawConfig, contextName)
		// if clusterName != nil {
		// 	clusterID = *clusterName
		// }
		klog.Errorf("Invalid ClusterID")
		return fmt.Errorf("invalid ClusterID")
	}

	if valid, err := isValidClusterID(joinConfig.ClusterID); !valid {
		klog.Errorf("Cluster ID invalid: %v", err)
		return err
	}

	_, failedRequirements, err := cmdVersion.CheckRequirements(r.Config)
	// We display failed requirements even if an error occurred
	if len(failedRequirements) > 0 {
		klog.Info("The target cluster fails to meet Submariner's requirements:")
		for i := range failedRequirements {
			klog.Infof("* %s", (failedRequirements)[i])
		}
		return fmt.Errorf("the target cluster fails to meet Submariner's requirements")
	}
	if err != nil {
		klog.Errorf("Unable to check all requirements: %v", err)
		return err
	}
	if brokerInfo.IsConnectivityEnabled() && joinConfig.LabelGateway {
		if err := r.HandleNodeLabels(); err != nil {
			klog.Errorf("Unable to set the gateway node up: %v", err)
			return err
		}
	}

	klog.Info("Discovering network details")
	networkDetails, err := r.GetNetworkDetails()
	if err != nil {
		klog.Errorf("Error get network details: %v", err)
		return err
	}
	serviceCIDR, serviceCIDRautoDetected, err := getServiceCIDR(joinConfig.ServiceCIDR, networkDetails)
	if err != nil {
		klog.Errorf("Error determining the service CIDR: %v", err)
		return err
	}
	clusterCIDR, clusterCIDRautoDetected, err := getPodCIDR(joinConfig.ClusterCIDR, networkDetails)
	if err != nil {
		klog.Errorf("Error determining the pod CIDR: %v", err)
		return err
	}

	brokerCluster, err := brokerInfo.GetBrokerAdministratorCluster()
	if err != nil {
		klog.Errorf("unable to get broker cluster client: %v", err)
		return err
	}
	brokerNamespace := string(brokerInfo.ClientToken.Data["namespace"])

	netconfig := globalnet.Config{
		ClusterID:               joinConfig.ClusterID,
		ServiceCIDR:             serviceCIDR,
		ServiceCIDRAutoDetected: serviceCIDRautoDetected,
		ClusterCIDR:             clusterCIDR,
		ClusterCIDRAutoDetected: clusterCIDRautoDetected,
		GlobalnetCIDR:           joinConfig.GlobalnetCIDR,
		GlobalnetClusterSize:    joinConfig.GlobalnetClusterSize,
	}
	if brokerInfo.IsGlobalnetEnabled() {
		if err = r.AllocateAndUpdateGlobalCIDRConfigMap(brokerCluster.GetClient(), brokerCluster.GetAPIReader(), instance, brokerNamespace, &netconfig); err != nil {
			klog.Errorf("Error Discovering multi cluster details: %v", err)
			return err
		}
	}

	klog.Info("Deploying the Submariner operator")
	if err = submarinerop.Ensure(r.Client, r.Config, true); err != nil {
		klog.Errorf("Error deploying the operator: %v", err)
		return err
	}
	klog.Info("Creating SA for cluster")
	clienttoken, err = broker.CreateSAForCluster(brokerCluster.GetClient(), brokerCluster.GetAPIReader(), joinConfig.ClusterID)
	if err != nil {
		klog.Errorf("Error creating SA for cluster: %v", err)
		return err
	}
	if brokerInfo.IsConnectivityEnabled() {
		klog.Info("Deploying Submariner")
		submarinerSpec, err := populateSubmarinerSpec(instance, brokerInfo, netconfig)
		if err != nil {
			return err
		}
		if err = submarinercr.Ensure(r.Client, consts.SubmarinerOperatorNamespace, submarinerSpec); err != nil {
			klog.Errorf("Submariner deployment failed: %v", err)
			return err
		}
		klog.Info("Submariner is up and running")
	} else if brokerInfo.IsServiceDiscoveryEnabled() {
		klog.Info("Deploying service discovery only")
		serviceDiscoverySpec, err := populateServiceDiscoverySpec(instance, brokerInfo)
		if err != nil {
			return err
		}
		if err = servicediscoverycr.Ensure(r.Client, consts.SubmarinerOperatorNamespace, serviceDiscoverySpec); err != nil {
			klog.Errorf("Service discovery deployment failed: %v", err)
			return err
		}
		klog.Info("Service discovery is up and running")
	}
	return nil
}

func (r *KnitnetReconciler) AllocateAndUpdateGlobalCIDRConfigMap(c client.Client, reader client.Reader, instance *operatorv1alpha1.Knitnet, brokerNamespace string,
	netconfig *globalnet.Config) error {
	joinConfig := instance.Spec.JoinConfig
	klog.Info("Discovering multi cluster details")
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		globalnetInfo, globalnetConfigMap, err := globalnet.GetGlobalNetworks(reader, brokerNamespace)
		if err != nil {
			klog.Errorf("error reading Global network details on Broker: %v", err)
			return err
		}

		netconfig.GlobalnetCIDR, err = globalnet.ValidateGlobalnetConfiguration(globalnetInfo, *netconfig)
		if err != nil {
			klog.Errorf("error validating Globalnet configuration: %v", err)
			return err
		}

		if globalnetInfo.GlobalnetEnabled {
			netconfig.GlobalnetCIDR, err = globalnet.AssignGlobalnetIPs(globalnetInfo, *netconfig)
			if err != nil {
				klog.Errorf("error assigning Globalnet IPs: %v", err)
				return err
			}

			if globalnetInfo.GlobalCidrInfo[joinConfig.ClusterID] == nil ||
				globalnetInfo.GlobalCidrInfo[joinConfig.ClusterID].GlobalCIDRs[0] != netconfig.GlobalnetCIDR {
				var newClusterInfo broker.ClusterInfo
				newClusterInfo.ClusterID = joinConfig.ClusterID
				newClusterInfo.GlobalCidr = []string{netconfig.GlobalnetCIDR}

				return broker.UpdateGlobalnetConfigMap(c, brokerNamespace, globalnetConfigMap, newClusterInfo)
			}
		}
		return err
	})
	return retryErr
}

func (r *KnitnetReconciler) GetNetworkDetails() (*network.ClusterNetwork, error) {
	dynClient, err := dynamic.NewForConfig(r.Config)
	if err != nil {
		return nil, err
	}

	networkDetails, err := network.Discover(dynClient, r.Client, consts.SubmarinerOperatorNamespace)
	if err != nil {
		klog.Errorf("Error trying to discover network details: %v", err)
	} else if networkDetails != nil {
		networkDetails.Show()
	}
	return networkDetails, nil
}

func getPodCIDR(clusterCIDR string, nd *network.ClusterNetwork) (cidrType string, autodetected bool, err error) {
	if clusterCIDR != "" {
		if nd != nil && len(nd.PodCIDRs) > 0 && nd.PodCIDRs[0] != clusterCIDR {
			klog.Warningf("Your provided cluster CIDR for the pods (%s) does not match discovered (%s)",
				clusterCIDR, nd.PodCIDRs[0])
		}
		return clusterCIDR, false, nil
	} else if nd != nil && len(nd.PodCIDRs) > 0 {
		return nd.PodCIDRs[0], true, nil
	}
	return "", true, fmt.Errorf("not found invalidate cluster CIDR")
}

func getServiceCIDR(serviceCIDR string, nd *network.ClusterNetwork) (cidrType string, autodetected bool, err error) {
	if serviceCIDR != "" {
		if nd != nil && len(nd.ServiceCIDRs) > 0 && nd.ServiceCIDRs[0] != serviceCIDR {
			klog.Warningf("Your provided service CIDR (%s) does not match discovered (%s)",
				serviceCIDR, nd.ServiceCIDRs[0])
		}
		return serviceCIDR, false, nil
	} else if nd != nil && len(nd.ServiceCIDRs) > 0 {
		return nd.ServiceCIDRs[0], true, nil
	}
	return "", true, fmt.Errorf("not found invalidate service CIDR")
}

func isValidClusterID(clusterID string) (bool, error) {
	// Make sure the clusterid is a valid DNS-1123 string
	if match, _ := regexp.MatchString("^[a-z0-9][a-z0-9.-]*[a-z0-9]$", clusterID); !match {
		return false, fmt.Errorf("cluster IDs must be valid DNS-1123 names, with only lowercase alphanumerics,\n"+
			"'.' or '-' (and the first and last characters must be alphanumerics).\n"+
			"%s doesn't meet these requirements", clusterID)
	}
	return true, nil
}

func populateSubmarinerSpec(instance *operatorv1alpha1.Knitnet, brokerInfo *broker.BrokerInfo, netconfig globalnet.Config) (*submariner.SubmarinerSpec, error) {
	joinConfig := instance.Spec.JoinConfig
	brokerURL := brokerInfo.BrokerURL
	if idx := strings.Index(brokerURL, "://"); idx >= 0 {
		// Submariner doesn't work with a schema prefix
		brokerURL = brokerURL[(idx + 3):]
	}

	// if our network discovery code was capable of discovering those CIDRs
	// we don't need to explicitly set it in the operator
	crServiceCIDR := ""
	if !netconfig.ServiceCIDRAutoDetected {
		crServiceCIDR = netconfig.ServiceCIDR
	}

	crClusterCIDR := ""
	if !netconfig.ClusterCIDRAutoDetected {
		crClusterCIDR = netconfig.ClusterCIDR
	}
	// customDomains := ""
	if joinConfig.CustomDomains == nil && brokerInfo.CustomDomains != nil {
		joinConfig.CustomDomains = *brokerInfo.CustomDomains
	}
	imageOverrides, err := getImageOverrides(instance)
	if err != nil {
		return nil, err
	}
	submarinerSpec := &submariner.SubmarinerSpec{
		Repository:               getImageRepo(instance),
		Version:                  getImageVersion(instance),
		CeIPSecNATTPort:          joinConfig.NattPort,
		CeIPSecIKEPort:           joinConfig.IkePort,
		CeIPSecDebug:             joinConfig.IpsecDebug,
		CeIPSecForceUDPEncaps:    joinConfig.ForceUDPEncaps,
		CeIPSecPreferredServer:   joinConfig.PreferredServer,
		CeIPSecPSK:               base64.StdEncoding.EncodeToString(brokerInfo.IPSecPSK.Data["psk"]),
		BrokerK8sCA:              base64.StdEncoding.EncodeToString(brokerInfo.ClientToken.Data["ca.crt"]),
		BrokerK8sRemoteNamespace: string(brokerInfo.ClientToken.Data["namespace"]),
		BrokerK8sApiServerToken:  string(clienttoken.Data["token"]),
		BrokerK8sApiServer:       brokerURL,
		Broker:                   "k8s",
		NatEnabled:               joinConfig.NatTraversal,
		Debug:                    joinConfig.SubmarinerDebug,
		ClusterID:                joinConfig.ClusterID,
		ServiceCIDR:              crServiceCIDR,
		ClusterCIDR:              crClusterCIDR,
		Namespace:                consts.SubmarinerOperatorNamespace,
		CableDriver:              joinConfig.CableDriver,
		ServiceDiscoveryEnabled:  brokerInfo.IsServiceDiscoveryEnabled(),
		ImageOverrides:           imageOverrides,
	}
	if netconfig.GlobalnetCIDR != "" {
		submarinerSpec.GlobalCIDR = netconfig.GlobalnetCIDR
	}
	if joinConfig.CorednsCustomConfigMap != "" {
		namespace, name := getCustomCoreDNSParams(instance)
		submarinerSpec.CoreDNSCustomConfig = &submariner.CoreDNSCustomConfig{
			ConfigMapName: name,
			Namespace:     namespace,
		}
	}
	if brokerInfo.CustomDomains != nil && len(*brokerInfo.CustomDomains) > 0 {
		submarinerSpec.CustomDomains = *brokerInfo.CustomDomains
	}
	return submarinerSpec, nil
}

func getImageVersion(instance *operatorv1alpha1.Knitnet) string {
	version := instance.Spec.JoinConfig.ImageVersion

	if version == "" {
		version = versions.DefaultSubmarinerOperatorVersion
	}

	return version
}

func getImageRepo(instance *operatorv1alpha1.Knitnet) string {
	repo := instance.Spec.JoinConfig.Repository

	if repo == "" {
		repo = versions.DefaultRepo
	}

	return repo
}

func removeSchemaPrefix(brokerURL string) string {
	if idx := strings.Index(brokerURL, "://"); idx >= 0 {
		// Submariner doesn't work with a schema prefix
		brokerURL = brokerURL[(idx + 3):]
	}

	return brokerURL
}

func populateServiceDiscoverySpec(instance *operatorv1alpha1.Knitnet, brokerInfo *broker.BrokerInfo) (*submariner.ServiceDiscoverySpec, error) {
	brokerURL := removeSchemaPrefix(brokerInfo.BrokerURL)
	joinConfig := instance.Spec.JoinConfig
	var customDomains []string
	if joinConfig.CustomDomains == nil && brokerInfo.CustomDomains != nil {
		customDomains = *brokerInfo.CustomDomains
	}
	imageOverrides, err := getImageOverrides(instance)
	if err != nil {
		return nil, err
	}
	serviceDiscoverySpec := submariner.ServiceDiscoverySpec{
		Repository:               joinConfig.Repository,
		Version:                  joinConfig.ImageVersion,
		BrokerK8sCA:              base64.StdEncoding.EncodeToString(brokerInfo.ClientToken.Data["ca.crt"]),
		BrokerK8sRemoteNamespace: string(brokerInfo.ClientToken.Data["namespace"]),
		BrokerK8sApiServerToken:  string(clienttoken.Data["token"]),
		BrokerK8sApiServer:       brokerURL,
		Debug:                    joinConfig.SubmarinerDebug,
		ClusterID:                joinConfig.ClusterID,
		Namespace:                consts.SubmarinerOperatorNamespace,
		ImageOverrides:           imageOverrides,
		GlobalnetEnabled:         brokerInfo.IsGlobalnetEnabled(),
	}

	if joinConfig.CorednsCustomConfigMap != "" {
		namespace, name := getCustomCoreDNSParams(instance)
		serviceDiscoverySpec.CoreDNSCustomConfig = &submariner.CoreDNSCustomConfig{
			ConfigMapName: name,
			Namespace:     namespace,
		}
	}

	if len(customDomains) > 0 {
		serviceDiscoverySpec.CustomDomains = customDomains
	}
	return &serviceDiscoverySpec, nil
}

func getImageOverrides(instance *operatorv1alpha1.Knitnet) (map[string]string, error) {
	joinConfig := instance.Spec.JoinConfig
	if len(joinConfig.ImageOverrideArr) > 0 {
		imageOverrides := make(map[string]string)
		for _, s := range joinConfig.ImageOverrideArr {
			key := strings.Split(s, "=")[0]
			if invalidImageName(key) {
				klog.Errorf("Invalid image name %s provided. Please choose from %q", key, names.ValidImageNames)
				return nil, fmt.Errorf("invalid image name %s provided. Please choose from %q", key, names.ValidImageNames)
			}
			value := strings.Split(s, "=")[1]
			imageOverrides[key] = value
		}
		return imageOverrides, nil
	}
	return nil, nil
}

func invalidImageName(key string) bool {
	for _, name := range names.ValidImageNames {
		if key == name {
			return false
		}
	}
	return true
}

func isValidCustomCoreDNSConfig(instance *operatorv1alpha1.Knitnet) error {
	corednsCustomConfigMap := instance.Spec.JoinConfig.CorednsCustomConfigMap
	if corednsCustomConfigMap != "" && strings.Count(corednsCustomConfigMap, "/") > 1 {
		klog.Error("coredns-custom-configmap should be in <namespace>/<name> format, namespace is optional")
		return fmt.Errorf("coredns-custom-configmap should be in <namespace>/<name> format, namespace is optional")
	}
	return nil
}

func getCustomCoreDNSParams(instance *operatorv1alpha1.Knitnet) (namespace, name string) {
	corednsCustomConfigMap := instance.Spec.JoinConfig.CorednsCustomConfigMap
	if corednsCustomConfigMap != "" {
		name = corednsCustomConfigMap
		paramList := strings.Split(corednsCustomConfigMap, "/")
		if len(paramList) > 1 {
			namespace = paramList[0]
			name = paramList[1]
		}
	}
	return namespace, name
}

func (r *KnitnetReconciler) HandleNodeLabels() error {
	const submarinerGatewayLabel = "submariner.io/gateway"
	const trueLabel = "true"
	selector, err := labels.Parse("submariner.io/gateway=true")
	if err != nil {
		return err
	}
	opts := &client.ListOptions{
		LabelSelector: selector,
	}
	nodes := &v1.NodeList{}
	if err := r.Client.List(context.TODO(), nodes, opts); err != nil {
		return err
	}
	if len(nodes.Items) > 0 {
		klog.Infof("* There are %d labeled nodes in the cluster:", len(nodes.Items))
		for _, node := range nodes.Items {
			klog.Infof("  - %s", node.GetName())
		}
	} else {
		node, err := r.getWorkerNodeForGateway()
		if err != nil {
			return err
		}
		if node == nil {
			klog.Info("* No worker node found to label as the gateway")
		} else {
			if err = r.addLabelsToNode(node.GetName(), map[string]string{submarinerGatewayLabel: trueLabel}); err != nil {
				klog.Errorf("Error labeling the gateway node: %v", err)
				return err
			}
		}
	}
	return nil
}
func (r *KnitnetReconciler) getWorkerNodeForGateway() (*v1.Node, error) {
	// List the worker nodes and select one
	workerNodes := &v1.NodeList{}
	workerSelector, err := labels.Parse("node-role.kubernetes.io/worker")
	if err != nil {
		klog.Errorf("Parse node label failed: %v", err)
		return nil, err
	}
	workerOpts := &client.ListOptions{
		LabelSelector: workerSelector,
	}

	if err := r.Client.List(context.TODO(), workerNodes, workerOpts); err != nil {
		klog.Errorf("List worker node failed: %v", err)
		return nil, err
	}
	if len(workerNodes.Items) == 0 {
		// In some deployments (like KIND), worker nodes are not explicitly labeled. So list non-master nodes.
		workerSelector, err := labels.Parse("!node-role.kubernetes.io/master")
		if err != nil {
			klog.Errorf("Parse node label failed: %v", err)
			return nil, err
		}
		workerOpts := &client.ListOptions{
			LabelSelector: workerSelector,
		}

		if err := r.Client.List(context.TODO(), workerNodes, workerOpts); err != nil {
			klog.Errorf("List non-master node failed: %v", err)
			return nil, err
		}
		if len(workerNodes.Items) == 0 {
			return nil, fmt.Errorf("not found any valid worker node for label")
		}
	}
	// Return the first node
	return &workerNodes.Items[0], nil
}

// this function was sourced from:
// https://github.com/kubernetes/kubernetes/blob/a3ccea9d8743f2ff82e41b6c2af6dc2c41dc7b10/test/utils/density_utils.go#L36
func (r *KnitnetReconciler) addLabelsToNode(nodeName string, labelsToAdd map[string]string) error {
	var tokens = make([]string, 0, len(labelsToAdd))
	for k, v := range labelsToAdd {
		tokens = append(tokens, fmt.Sprintf("\"%s\":\"%s\"", k, v))
	}

	labelString := "{" + strings.Join(tokens, ",") + "}"
	patch := []byte(fmt.Sprintf(`{"metadata":{"labels":%v}}`, labelString))

	var lastErr error
	err := wait.ExponentialBackoff(nodeLabelBackoff, func() (bool, error) {
		node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: nodeName}}
		lastErr := r.Client.Patch(context.TODO(), node, client.RawPatch(types.StrategicMergePatchType, patch))
		if lastErr != nil {
			if !errors.IsConflict(lastErr) {
				return false, lastErr
			}
			return false, nil
		} else {
			return true, nil
		}
	})

	if err == wait.ErrWaitTimeout {
		return lastErr
	}
	return err
}
