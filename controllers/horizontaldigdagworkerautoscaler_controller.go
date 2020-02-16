/*

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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	horizontalpodautoscalersautoscalingv1 "digdag-worker-crd/api/v1"
)

// HorizontalDigdagWorkerAutoscalerReconciler reconciles a HorizontalDigdagWorkerAutoscaler object
type HorizontalDigdagWorkerAutoscalerReconciler struct {
	client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=horizontalpodautoscalers.autoscaling.digdag-worker-crd,resources=horizontaldigdagworkerautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=horizontalpodautoscalers.autoscaling.digdag-worker-crd,resources=horizontaldigdagworkerautoscalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps;extensions,resources=deployments,verbs=get;list;watch;create;update;patch

func (r *HorizontalDigdagWorkerAutoscalerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	// featch list of HorizontalDigdagWorkerAutoscaler
	horizontalDigdagWorkerAutoscalers := &horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscalerList{}
	err := r.List(context.Background(), &client.ListOptions{}, horizontalDigdagWorkerAutoscalers)
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, horizontalDigdagWorkerAutoscalerItem := range horizontalDigdagWorkerAutoscalers.Items {
		// Obtain the target HorizontalDigdagWorkerAutoscaler from the MetaData of HorizontalDigdagWorkerAutoscaler
		instance := &horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscalerSpec{}
		err = req.Get(context.Background(), types.NamespacedName{
			Name:      horizontalDigdagWorkerAutoscalerItem.ObjectMeta.Name,
			Namespace: horizontalDigdagWorkerAutoscalerItem.ObjectMeta.Namespace,
		}, instance)

		if err != nil {
			if !errors.IsNotFound(err) {
				return reconcile.Result{}, err
			}
		}

		//TODO Obtain digdag task queue info from HorizontalDigdagWorkerAutoscaler's configure

		//TODO Obtain the number of pods (replica) of Deployment linked to HorizontalDigdagWorkerAutoscaler

		//TODO Update the number of deployment pods according to the task queue

	}
	return ctrl.Result{}, nil
}

// SetupWithManager registers this reconciler with the controller manager and
// starts watching Deployment.
func (r *HorizontalDigdagWorkerAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscaler{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
