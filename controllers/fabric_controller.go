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
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1alpha1 "github.com/tkestack/cluster-fabric-operator/api/v1alpha1"
	consts "github.com/tkestack/cluster-fabric-operator/controllers/ensures"
	"github.com/tkestack/cluster-fabric-operator/controllers/ensures/broker"
)

// FabricReconciler reconciles a Fabric object
type FabricReconciler struct {
	client.Client
	client.Reader
	*rest.Config
	Scheme *runtime.Scheme
}

const (
	BrokerAction = "broker"
	JoinAction   = "join"
	AllAction    = "all"
)

//+kubebuilder:rbac:groups=operator.tkestack.io,resources=fabrics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.tkestack.io,resources=fabrics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.tkestack.io,resources=fabrics/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Fabric object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *FabricReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, err error) {
	klog.Infof("Start reconciling Fabric: %s", req.NamespacedName)
	instance := &operatorv1alpha1.Fabric{}

	if err := r.Client.Get(context.TODO(), req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	originalInstance := instance.DeepCopy()
	// Always attempt to patch the status after each reconciliation.
	defer func() {
		if err != nil {
			instance.Status.Phase = operatorv1alpha1.PhaseFailed
		} else {
			instance.Status.Phase = operatorv1alpha1.PhaseRunning
		}
		if reflect.DeepEqual(originalInstance.Status, instance.Status) {
			return
		}
		if updateErr := r.Status().Update(ctx, instance, &client.UpdateOptions{}); updateErr != nil {
			klog.Errorf("Update status failed, err: %v", updateErr)
		}
	}()

	// Deploy submeriner broker
	if instance.Spec.Action == BrokerAction || instance.Spec.Action == AllAction {
		klog.Info("Deploy submeriner broker")
		if err := r.DeploySubmerinerBroker(instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Join managed cluster to submeriner borker
	if instance.Spec.Action == JoinAction || instance.Spec.Action == AllAction {
		klog.Info("Join managed cluster to submeriner broker")
		brokerInfo, err := broker.NewFromConfigMap(r.Client)
		if err != nil {
			return ctrl.Result{}, err
		}
		if err := r.JoinSubmarinerCluster(instance, brokerInfo); err != nil {
			return ctrl.Result{}, err
		}
	}
	klog.Infof("Finished reconciling Fabric: %s", req.NamespacedName)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FabricReconciler) SetupWithManager(mgr ctrl.Manager) error {
	cmPredicates := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			labels := e.Object.GetLabels()
			for labelKey := range labels {
				if labelKey == consts.FabricNameLabel {
					return true
				}
			}
			return false
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Fabric{}).
		Watches(
			&source.Kind{Type: &corev1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
				lables := obj.GetLabels()
				name, nameOk := lables[consts.FabricNameLabel]
				ns, namespaceOK := lables[consts.FabricNamespaceLabel]
				if nameOk && namespaceOK {
					return []reconcile.Request{
						{NamespacedName: types.NamespacedName{
							Name:      name,
							Namespace: ns,
						}},
					}
				}
				return nil
			}),
			builder.WithPredicates(cmPredicates),
		).
		Complete(r)
}
