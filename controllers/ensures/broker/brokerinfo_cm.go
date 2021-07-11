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
	"crypto/rand"
	"encoding/base64"
	"encoding/json"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/tkestack/knitnet-operator/api/v1alpha1"
	"github.com/tkestack/knitnet-operator/controllers/stringset"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/cluster"

	"github.com/tkestack/knitnet-operator/controllers/components"
	consts "github.com/tkestack/knitnet-operator/controllers/ensures"
)

type BrokerInfo struct {
	BrokerURL                   string     `json:"brokerURL"`
	ClientToken                 *v1.Secret `json:"clientToken,omitempty"`
	IPSecPSK                    *v1.Secret `json:"ipsecPSK,omitempty"`
	Components                  []string   `json:",omitempty"`
	CustomDomains               *[]string  `json:"customDomains,omitempty"`
	GlobalnetCIDRRange          string     `json:"globalnetCIDRRange,omitempty"`
	DefaultGlobalnetClusterSize uint       `json:"defaultGlobalnetClusterSize,omitempty"`
}

const ipsecPSKSecretName = "submariner-ipsec-psk"
const ipsecSecretLength = 48

func (data *BrokerInfo) SetComponents(componentSet stringset.Interface) {
	data.Components = componentSet.Elements()
}

func (data *BrokerInfo) GetComponents() stringset.Interface {
	return stringset.New(data.Components...)
}

func (data *BrokerInfo) IsConnectivityEnabled() bool {
	return data.GetComponents().Contains(components.Connectivity)
}

func (data *BrokerInfo) IsServiceDiscoveryEnabled() bool {
	return data.GetComponents().Contains(components.ServiceDiscovery)
}

func (data *BrokerInfo) IsGlobalnetEnabled() bool {
	return data.GetComponents().Contains(components.Globalnet)
}

func (data *BrokerInfo) ToString() (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(jsonBytes), nil
}

func NewFromString(str string) (*BrokerInfo, error) {
	data := &BrokerInfo{}
	bytes, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return data, json.Unmarshal(bytes, data)
}

func (data *BrokerInfo) WriteConfigMap(c client.Client, instance *operatorv1alpha1.Knitnet) error {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      consts.SubmarinerBrokerInfo,
			Namespace: consts.SubmarinerBrokerNamespace,
		},
	}
	labels := make(map[string]string)
	labels[consts.KnitnetNameLabel] = instance.GetName()
	labels[consts.KnitnetNamespaceLabel] = instance.GetNamespace()

	or, err := ctrl.CreateOrUpdate(context.TODO(), c, cm, func() error {
		dataStr, err := data.ToString()
		if err != nil {
			return err
		}
		cm.ObjectMeta.Labels = labels
		cm.Data = map[string]string{"brokerInfo": dataStr}
		return nil
	})
	if err != nil {
		return err
	}
	klog.Infof("Configmap %s %s", consts.SubmarinerBrokerInfo, or)
	return nil
}

func NewFromConfigMap(c client.Client) (*BrokerInfo, error) {
	cm := &v1.ConfigMap{}
	cmKey := types.NamespacedName{Name: consts.SubmarinerBrokerInfo, Namespace: consts.SubmarinerBrokerNamespace}
	if err := c.Get(context.TODO(), cmKey, cm); err != nil {
		return nil, err
	}
	return NewFromString(cm.Data["brokerInfo"])
}

func NewFromCluster(c client.Client, restConfig *rest.Config) (*BrokerInfo, error) {
	brokerInfo := &BrokerInfo{}
	var err error
	brokerInfo.ClientToken, err = GetClientTokenSecret(c, consts.SubmarinerBrokerNamespace, SubmarinerBrokerAdminSA)
	if err != nil {
		return nil, err
	}
	brokerInfo.IPSecPSK, err = newIPSECPSKSecret()
	if err != nil {
		return nil, err
	}
	brokerInfo.BrokerURL = restConfig.Host + restConfig.APIPath
	return brokerInfo, err
}

func CreateBrokerInfoConfigMap(c client.Client, restConfig *rest.Config, instance *operatorv1alpha1.Knitnet) error {
	klog.Info("Create or update broker info configmap")
	brokerInfo, err := NewFromCluster(c, restConfig)
	if err != nil {
		return err
	}
	brokerConfig := instance.Spec.BrokerConfig
	brokerInfo.GlobalnetCIDRRange = brokerConfig.GlobalnetCIDRRange
	brokerInfo.DefaultGlobalnetClusterSize = brokerConfig.DefaultGlobalnetClusterSize

	if len(brokerConfig.DefaultCustomDomains) > 0 {
		brokerInfo.CustomDomains = &brokerConfig.DefaultCustomDomains
	}

	if err := brokerInfo.WriteConfigMap(c, instance); err != nil {
		return err
	}
	return nil
}

func (data *BrokerInfo) GetBrokerAdministratorCluster() (cluster.Cluster, error) {
	config := data.GetBrokerAdministratorConfig()
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	return cluster.New(config, func(clusterOptions *cluster.Options) {
		clusterOptions.Scheme = scheme
	})
}

func (data *BrokerInfo) GetBrokerAdministratorConfig() *rest.Config {
	tlsClientConfig := rest.TLSClientConfig{}
	if len(data.ClientToken.Data["ca.crt"]) != 0 {
		tlsClientConfig.CAData = data.ClientToken.Data["ca.crt"]
	}
	bearerToken := data.ClientToken.Data["token"]
	restConfig := rest.Config{
		Host:            data.BrokerURL,
		TLSClientConfig: tlsClientConfig,
		BearerToken:     string(bearerToken),
	}
	return &restConfig
}

// generateRandomPSK returns securely generated n-byte array.
func generateRandomPSK(n int) ([]byte, error) {
	psk := make([]byte, n)
	_, err := rand.Read(psk)
	return psk, err
}

func newIPSECPSKSecret() (*v1.Secret, error) {
	psk, err := generateRandomPSK(ipsecSecretLength)
	if err != nil {
		return nil, err
	}

	pskSecretData := make(map[string][]byte)
	pskSecretData["psk"] = psk

	pskSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: ipsecPSKSecretName,
		},
		Data: pskSecretData,
	}

	return pskSecret, nil
}
