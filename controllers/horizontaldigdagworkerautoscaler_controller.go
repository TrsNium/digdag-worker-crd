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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hpav1 "digdag-worker-crd/api/v1"
)

// HorizontalDigdagWorkerAutoscalerReconciler reconciles a HorizontalDigdagWorkerAutoscaler object
type HorizontalDigdagWorkerAutoscalerReconciler struct {
	Client                   client.Client
	Log                      logr.Logger
	Scheme                   *runtime.Scheme
	DigdagWorkerScaleManager *DigdagWorkerScaleManager
}

// +kubebuilder:rbac:groups=horizontalpodautoscalers.autoscaling.digdag-worker-crd,resources=horizontaldigdagworkerautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=horizontalpodautoscalers.autoscaling.digdag-worker-crd,resources=horizontaldigdagworkerautoscalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=configmap,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=services;serviceaccounts;secrets,verbs=get;list;watch;create;update;delete;patch
func (r *HorizontalDigdagWorkerAutoscalerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("HorizontalDigdagWorkerAutoscaler", req.NamespacedName)

	// fetch list of HorizontalDigdagWorkerAutoscaler
	horizontalDigdagWorkerAutoscalers := &hpav1.HorizontalDigdagWorkerAutoscalerList{}
	if err := r.Client.List(ctx, horizontalDigdagWorkerAutoscalers, &client.ListOptions{}); err != nil {
		log.Error(err, "failed to get HorizontalDigdagWorkerAutoscaler resource")
		// Ignore NotFound errors as they will be retried automatically if the
		// resource is created in future.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	for _, horizontalDigdagWorkerAutoscaler := range horizontalDigdagWorkerAutoscalers.Items {
		if !r.DigdagWorkerScaleManager.IsManaged(horizontalDigdagWorkerAutoscaler) {
			r.Log.Info("Try to regist DigdagWorkerScaleManager")
			err := r.DigdagWorkerScaleManager.Manage(horizontalDigdagWorkerAutoscaler)
			if err != nil {
				log.Error(err, "failed to manage new digdagWorkerScalers")
			}
			continue
		}

		r.Log.Info("Check to resource is updated.")
		if r.DigdagWorkerScaleManager.IsUpdated(horizontalDigdagWorkerAutoscaler) {
			err := r.DigdagWorkerScaleManager.Update(horizontalDigdagWorkerAutoscaler)
			if err != nil {
				log.Error(err, "failed to update DigdagWorkerScaleManager")
				continue
			}
		}
	}

	// finaly gc not used DigdagWorkerScaler
	r.DigdagWorkerScaleManager.GCNotUsed(horizontalDigdagWorkerAutoscalers)
	return ctrl.Result{}, nil
}

// SetupWithManager registers this reconciler with the controller manager and
// starts watching Deployment.
func (r *HorizontalDigdagWorkerAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hpav1.HorizontalDigdagWorkerAutoscaler{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
