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
	"database/sql"
	_ "github.com/lib/pq"
	"math"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
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
	ctx := context.Background()
	log := r.Log.WithValues("HorizontalDigdagWorkerAutoscaler", req.NamespacedName)

	// featch list of HorizontalDigdagWorkerAutoscaler
	horizontalDigdagWorkerAutoscaler := &horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscaler{}
	if err := r.Client.Get(ctx, req.NamespacedName, &horizontalDigdagWorkerAutoscalers); err != nil {
		log.Error(err, "failed to get HorizontalDigdagWorkerAutoscaler resource")
		// Ignore NotFound errors as they will be retried automatically if the
		// resource is created in future.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("checking if an existing Deployment exists for this resource")

	// Get Deployment associated with HorizontalDigdagWorkerAutoscaler
	deployment := appsv1.Deployment{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: horizontalDigdagWorkerAutoscaler.Namespace, Name: horizontalDigdagWorkerAutoscaler.Spec.ScaleTargetDeployment}, &deployment)
	if errors.IsNotFound(err) {
		log.Info("Deployment associated with HorizontalDigdagWorkerAutoscaler was not found")
		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Error(err, "failed to get Deployment for MyKind resource")
		return ctrl.Result{}, err
	}

	// Obtain digdag task queue info from HorizontalDigdagWorkerAutoscaler's configure
	queuedTaskNum, err := getQueuedTaskNum(horizontalDigdagWorkerAutoscaler.Spec)
	if err != nil {
		log.Error(err, "failed to get queuedTaskNum")
		return ctrl.Result{}, err
	}

	// TODO: Set replicas to 1 if workflow is not running
	if queuedTotalTaskNum == 0 {
		// Set replicas to 1 because there are no tasks to execute
		log.Info("Digdag is idling now")

		expectedReplicas := int32(1)
		deployment.Spec.Replicas = &expectedReplicas
		if err := r.Client.Update(ctx, &deployment); err != nil {
			log.Error(err, "failed to Deployment update replica count")
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(&horizontalDigdagWorkerAutoscaler, core.EventTypeNormal, "Scaled", "Scaled deployment %q to %d replicas", deployment.Name, expectedReplicas)
		return ctrl.Result{}, nil
	} else {
		runningTaskNum, err := getRunningTaskNum()
		if err != nil {
			log.Error(err, "failed to get planingTaskNum")
			return ctrl.Result{}, err
		}

		// Obtain replica of Deployment linked to HorizontalDigdagWorkerAutoscaler
		currentReplicas := *deployment.Spec.Replicas

		// Update the number of deployment pods according to the task queue
		digdagWorkerMaxTaskThreads := horizontalDigdagWorkerAutoscaler.Spec.DigdagWorkerMaxTaskThreads
		digdagTotalTaskThreads := currentReplicas * digdagWorkerMaxTaskThreads

		// NOTE
		// Tasks that are not running on any node will be in the running state,
		// So subtracting digdagTotalTaskThreads from all running tasks will give you the number of surplus tasks
		surplusTaskNum := runningTaskNum - digdagTotalTaskThreads
		if surplusTaskNum > 0 {
			additionalReplicas := int32(math.math.Ceil(surplusTaskNum / digdagWorkerMaxTaskThreads))
			newReplicas := currentReplicas + additionalReplicas

			deployment.Spec.Replicas = &newReplicas
			if err := r.Client.Update(ctx, &deployment); err != nil {
				log.Error(err, "failed to Deployment update replica count")
				return ctrl.Result{}, err
			}

			r.Recorder.Eventf(&horizontalDigdagWorkerAutoscaler, core.EventTypeNormal, "Scaled", "Scaled deployment %q to %d replicas", deployment.Name, newReplicas)
			return ctrl.Result{}, nil
		}
	}
	// No need to change the number of replicas
	return ctrl.Result{}, nil
}

// select count(*) from tasks;
func getQueuedTaskNum(spec horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscalerSpec) (error, int32) {

}

// select count(*) from tasks where task_type = 0 and state = 2;
func getRunningTaskNum(spec horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscalerSpec) (error, int32) {

}

// SetupWithManager registers this reconciler with the controller manager and
// starts watching Deployment.
func (r *HorizontalDigdagWorkerAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscaler{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
