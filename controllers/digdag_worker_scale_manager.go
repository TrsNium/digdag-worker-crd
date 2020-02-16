package controllers

import (
	"fmt"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	horizontalpodautoscalersautoscalingv1 "digdag-worker-crd/api/v1"
)

type DigdagWorkerScaleManager struct {
	client              client.Client
	log                 logr.Logger
	digdagWorkerScalers map[string]DigdagWorkerScaler
}

func (r *DigdagWorkerScaleManager) IsManaged(horizontalDigdagWorkerAutoscaler horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscaler) bool {
	objectMeta := horizontalDigdagWorkerAutoscaler.ObjectMeta
	key := fmt.Sprintf("%s-%s", objectMeta.Namespace, objectMeta.Name)
	_, isManaged := r.digdagWorkerScalers[key]
	return isManaged
}

func (r *DigdagWorkerScaleManager) IsUpdated(horizontalDigdagWorkerAutoscaler horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscaler) bool {
	objectMeta := horizontalDigdagWorkerAutoscaler.ObjectMeta
	key := fmt.Sprintf("%s-%s", objectMeta.Namespace, objectMeta.Name)
	digdagWorkerScaler, _ := r.digdagWorkerScalers[key]
	return digdagWorkerScaler.Equal(horizontalDigdagWorkerAutoscaler)
}

func (r *DigdagWorkerScaleManager) Manage(horizontalDigdagWorkerAutoscaler horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscaler) error {
	objectMeta := horizontalDigdagWorkerAutoscaler.ObjectMeta
	key := fmt.Sprintf("%s-%s", objectMeta.Namespace, objectMeta.Name)
	digdagWorkerScaler, err := NewDigdagWorkerScaler(r.client, r.log, horizontalDigdagWorkerAutoscaler)
	if err != nil {
		return err
	}
	r.digdagWorkerScalers[key] = digdagWorkerScaler
}

func (r *DigdagWorkerScaleManager) Update(horizontalDigdagWorkerAutoscaler horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscaler) error {
	objectMeta := horizontalDigdagWorkerAutoscaler.ObjectMeta
	key := fmt.Sprintf("%s-%s", objectMeta.Namespace, objectMeta.Name)
	digdagWorkerScaler, _ := r.digdagWorkerScalers[key]
	err := digdagWorkerScaler.Update(horizontalDigdagWorkerAutoscaler)
	return err
}

func (r *DigdagWorkerScaleManager) gc(digdagWorkerScalersKey string) {
	digdagWorkerScaler, _ := r.digdagWorkerScalers[digdagWorkerScalersKey]
	digdagWorkerScaler.GC()
}

func (r *DigdagWorkerScaleManager) GCNotUsed(horizontalDigdagWorkerAutoscalers horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscalerList) error {
	keys := []string{}
	for _, horizontalDigdagWorkerAutoscaler := range horizontalDigdagWorkerAutoscalers {
		objectMeta := horizontalDigdagWorkerAutoscaler.ObjectMeta
		key := fmt.Sprintf("%s-%s", objectMeta.Namespace, objectMeta.Name)
		keys = append(keys, key)
	}

	digdagWorkerScalersKeys := keys(r.digdagWorkerScalers)
	for _, digdagWorkerScalersKey := range digdagWorkerScalersKeys {
		if !contains(keys, digdagWorkerScalersKey) {
			r.gc(digdagWorkerScalersKey)
		}
	}
}

func keys(m map[string]DigdagWorkerScaler) []string {
	ks := []string{}
	for k, _ := range m {
		ks = append(ks, k)
	}
	return ks
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}

func NewDigdagWorkerScaleManager(client client.Client, log logr.Logger) *DigdagWorkerScaleManager {
	dwsm := &DigdagWorkerScaleManager{
		client:              client,
		log:                 logr.Logger,
		digdagWorkerScalers: make(map[string]DigdagWorkerScaler),
	}
	return dwsm
}
