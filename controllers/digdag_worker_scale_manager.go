package controllers

import (
	"fmt"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hpav1 "digdag-worker-crd/api/v1"
)

type DigdagWorkerScaleManager struct {
	client              client.Client
	log                 logr.Logger
	digdagWorkerScalers map[string]DigdagWorkerScaler
}

func (r *DigdagWorkerScaleManager) IsManaged(horizontalDigdagWorkerAutoscaler hpav1.HorizontalDigdagWorkerAutoscaler) bool {
	dployment := horizontalDigdagWorkerAutoscaler.Spec.Deployment
	key := fmt.Sprintf("%s-%s", dployment.Namespace, dployment.Name)
	_, isManaged := r.digdagWorkerScalers[key]
	return isManaged
}

func (r *DigdagWorkerScaleManager) IsUpdated(horizontalDigdagWorkerAutoscaler hpav1.HorizontalDigdagWorkerAutoscaler) bool {
	dployment := horizontalDigdagWorkerAutoscaler.Spec.Deployment
	key := fmt.Sprintf("%s-%s", dployment.Namespace, dployment.Name)
	digdagWorkerScaler, _ := r.digdagWorkerScalers[key]
	return digdagWorkerScaler.Equal(horizontalDigdagWorkerAutoscaler)
}

func (r *DigdagWorkerScaleManager) Manage(horizontalDigdagWorkerAutoscaler hpav1.HorizontalDigdagWorkerAutoscaler) error {
	dployment := horizontalDigdagWorkerAutoscaler.Spec.Deployment
	key := fmt.Sprintf("%s-%s", dployment.Namespace, dployment.Name)
	digdagWorkerScaler, err := NewDigdagWorkerScaler(r.client, r.log, horizontalDigdagWorkerAutoscaler)
	if err != nil {
		return err
	}
	r.digdagWorkerScalers[key] = digdagWorkerScaler
	r.log.Info(fmt.Sprintf("Start to manage %s", key))
	return nil
}

func (r *DigdagWorkerScaleManager) Update(horizontalDigdagWorkerAutoscaler hpav1.HorizontalDigdagWorkerAutoscaler) error {
	dployment := horizontalDigdagWorkerAutoscaler.Spec.Deployment
	key := fmt.Sprintf("%s-%s", dployment.Namespace, dployment.Name)
	digdagWorkerScaler, _ := r.digdagWorkerScalers[key]
	err := digdagWorkerScaler.Update(horizontalDigdagWorkerAutoscaler)
	r.log.Info(fmt.Sprintf("%s is updated", key))
	return err
}

func (r *DigdagWorkerScaleManager) gc(digdagWorkerScalersKey string) {
	digdagWorkerScaler, _ := r.digdagWorkerScalers[digdagWorkerScalersKey]
	digdagWorkerScaler.GC()
}

func (r *DigdagWorkerScaleManager) GCNotUsed(horizontalDigdagWorkerAutoscalers *hpav1.HorizontalDigdagWorkerAutoscalerList) {
	keys := []string{}
	for _, horizontalDigdagWorkerAutoscaler := range horizontalDigdagWorkerAutoscalers.Items {
		dployment := horizontalDigdagWorkerAutoscaler.Spec.Deployment
		key := fmt.Sprintf("%s-%s", dployment.Namespace, dployment.Name)
		keys = append(keys, key)
	}

	digdagWorkerScalersKeys := r.keys(r.digdagWorkerScalers)
	for _, digdagWorkerScalersKey := range digdagWorkerScalersKeys {
		if !r.contains(keys, digdagWorkerScalersKey) {
			r.gc(digdagWorkerScalersKey)
		}
	}
}

func (r *DigdagWorkerScaleManager) keys(m map[string]DigdagWorkerScaler) []string {
	ks := []string{}
	for k, _ := range m {
		ks = append(ks, k)
	}
	return ks
}

func (r *DigdagWorkerScaleManager) contains(s []string, e string) bool {
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
		log:                 log,
		digdagWorkerScalers: make(map[string]DigdagWorkerScaler),
	}
	return dwsm
}
