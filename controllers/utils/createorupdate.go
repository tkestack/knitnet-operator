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

package utils

import (
	"context"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tkestack/knitnet-operator/controllers/embeddedyamls"
)

type CRDUpdater interface {
	Create(ctx context.Context, customResourceDefinition *apiextensionsv1.CustomResourceDefinition, opts metav1.CreateOptions) (*apiextensionsv1.CustomResourceDefinition, error)
	Update(ctx context.Context, customResourceDefinition *apiextensionsv1.CustomResourceDefinition, opts metav1.UpdateOptions) (*apiextensionsv1.CustomResourceDefinition, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apiextensionsv1.CustomResourceDefinition, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
}

func NewFromRestConfig(config *rest.Config) (CRDUpdater, error) {
	apiext, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating the api extensions client: %s", err)
	}
	return apiext.ApiextensionsV1().CustomResourceDefinitions(), nil
}

func CreateOrUpdateEmbeddedCRD(c client.Client, crdYaml string) error {
	crdName, err := embeddedyamls.GetObjectName(crdYaml)
	if err != nil {
		return err
	}
	crd := &apiextensionsv1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: crdName}}
	or, err := ctrl.CreateOrUpdate(context.TODO(), c, crd, func() error {
		return embeddedyamls.GetObject(crdYaml, crd)
	})
	if err != nil {
		klog.Errorf("Failed to %s CRD %s: %v", or, crd.GetName(), err)
		return err
	}
	klog.Infof("CRD %s %s", crd.GetName(), or)
	return nil
}
