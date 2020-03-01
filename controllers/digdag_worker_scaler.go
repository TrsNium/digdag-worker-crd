package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-logr/logr"
	_ "github.com/lib/pq"
	"github.com/robfig/cron"
	"math"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hpav1 "digdag-worker-crd/api/v1"
)

type DigdagWorkerScaler struct {
	client           client.Client
	logger           logr.Logger
	deployment       hpav1.Deployment
	postgresql       hpav1.Postgresql
	scaleIntervalSec int32
	cron             *cron.Cron
	db               *sql.DB
}

func (r *DigdagWorkerScaler) Equal(horizontalDigdagWorkerAutoscaler hpav1.HorizontalDigdagWorkerAutoscaler) bool {
	deployment := horizontalDigdagWorkerAutoscaler.Spec.Deployment
	postgresql := horizontalDigdagWorkerAutoscaler.Spec.Postgresql
	return reflect.DeepEqual(&r.deployment, &deployment) ||
		reflect.DeepEqual(&r.postgresql, &postgresql) ||
		r.scaleIntervalSec != 15
}

func (r *DigdagWorkerScaler) Update(horizontalDigdagWorkerAutoscaler hpav1.HorizontalDigdagWorkerAutoscaler) error {
	r.GC()

	deployment := horizontalDigdagWorkerAutoscaler.Spec.Deployment
	postgresql := horizontalDigdagWorkerAutoscaler.Spec.Postgresql
	db, err := createDB(
		r.client,
		r.logger,
		postgresql.Host,
		postgresql.Port,
		postgresql.Database,
		postgresql.User,
		postgresql.Password,
	)
	if err != nil {
		return err
	}

	r.deployment = deployment
	r.postgresql = postgresql
	r.scaleIntervalSec = 15
	r.db = db

	cron := cron.New()
	cron.AddFunc("*/15 * * * * *", r.scaleDigdagWorker)
	cron.Start()
	r.cron = cron
	return nil
}

func (r *DigdagWorkerScaler) getQueuedTaskNum() (int32, error) {
	var count int32
	if err := r.db.QueryRow("select count(*) from tasks;").Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *DigdagWorkerScaler) getRunningTaskNum() (int32, error) {
	var count int32
	if err := r.db.QueryRow("select count(*) from tasks where task_type = 0 and state = 2;").Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *DigdagWorkerScaler) scaleDigdagWorker() {
	ctx := context.Background()
	// Get Deployment associated with HorizontalDigdagWorkerAutoscaler
	deployment := appsv1.Deployment{}
	err := r.client.Get(ctx, client.ObjectKey{Namespace: r.deployment.Namespace, Name: r.deployment.Name}, &deployment)
	if errors.IsNotFound(err) {
		r.logger.Info("Deployment associated with HorizontalDigdagWorkerAutoscaler was not found")
		return
	}
	if err != nil {
		r.logger.Error(err, "failed to get Deployment for MyKind resource")
		return
	}

	// Obtain digdag task queue info from HorizontalDigdagWorkerAutoscaler's configure
	queuedTaskNum, err := r.getQueuedTaskNum()
	if err != nil {
		r.logger.Error(err, "failed to get queuedTaskNum")
		return
	}

	// If there are no queued tasks, set Replicas to 1.
	// If there are queued tasks, get the number of tasks that have not been executed and update the Replicas.
	if queuedTaskNum == 0 {
		// Set replicas to 1 because there are no tasks to execute
		r.logger.Info("Digdag is idling")

		var minReplicas *int32 = r.deployment.MinReplicas
		if minReplicas == nil {
			defaultMinReplicas := int32(1)
			minReplicas = &defaultMinReplicas
		}

		deployment.Spec.Replicas = minReplicas
		if err := r.client.Update(ctx, &deployment); err != nil {
			r.logger.Error(err, "failed to Deployment update replica count")
			return
		}

		return
	} else {
		runningTaskNum, err := r.getRunningTaskNum()
		if err != nil {
			r.logger.Error(err, "failed to get planingTaskNum")
			return
		}

		// Obtain replica of Deployment linked to HorizontalDigdagWorkerAutoscaler
		currentReplicas := *deployment.Spec.Replicas

		// Update the number of deployment pods according to the task queue
		digdagWorkerMaxTaskThreads := r.deployment.MaxTaskThreads
		digdagTotalTaskThreads := currentReplicas * digdagWorkerMaxTaskThreads

		// NOTE
		// Tasks that are not running on any node will be in the running state,
		// So subtracting digdagTotalTaskThreads from all running tasks will give you the number of surplus tasks
		surplusTaskNum := runningTaskNum - digdagTotalTaskThreads
		if surplusTaskNum > 0 {
			additionalReplicas := int32(math.Ceil(float64(surplusTaskNum) / float64(digdagWorkerMaxTaskThreads)))
			newReplicas := currentReplicas + additionalReplicas

			deployment.Spec.Replicas = &newReplicas
			r.logger.Info(fmt.Sprintf("Scale to %d", newReplicas))
			if err := r.client.Update(ctx, &deployment); err != nil {
				r.logger.Error(err, "failed to Deployment update replica count")
				return
			}
			return
		}
		r.logger.Info(fmt.Sprintf("Keep replicas %d", currentReplicas))
	}
}

func (r *DigdagWorkerScaler) GC() {
	r.logger.Info("GC")
	r.db.Close()
	r.cron.Stop()
}

func createDB(
	client client.Client,
	logr logr.Logger,
	hostRef hpav1.Ref,
	portRef hpav1.Ref,
	databaseRef hpav1.Ref,
	userRef hpav1.Ref,
	passwordRef hpav1.Ref,
) (*sql.DB, error) {

	host, err := hostRef.GetValue(client)
	if err != nil {
		logr.Error(err, "Could not read host value")
		return nil, err
	}

	port, err := portRef.GetValue(client)
	if err != nil {
		logr.Error(err, "Could not read port value")
		return nil, err
	}

	database, err := databaseRef.GetValue(client)
	if err != nil {
		logr.Error(err, "Could not read database value")
		return nil, err
	}

	user, err := userRef.GetValue(client)
	if err != nil {
		logr.Error(err, "Could not read user value")
		return nil, err
	}

	password, err := passwordRef.GetValue(client)
	if err != nil {
		logr.Error(err, "Could not read password value")
		return nil, err
	}

	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", host, port, database, user, password)
	return sql.Open("postgres", connStr)
}

func NewDigdagWorkerScaler(client client.Client, logr logr.Logger, horizontalDigdagWorkerAutoscaler hpav1.HorizontalDigdagWorkerAutoscaler) (DigdagWorkerScaler, error) {
	logr.Info("Create new DigdagWorkerScaler")
	postgresql := horizontalDigdagWorkerAutoscaler.Spec.Postgresql
	db, err := createDB(
		client,
		logr,
		postgresql.Host,
		postgresql.Port,
		postgresql.Database,
		postgresql.User,
		postgresql.Password,
	)
	if err != nil {
		return DigdagWorkerScaler{}, err
	}

	scaler := DigdagWorkerScaler{
		client:           client,
		logger:           logr,
		deployment:       horizontalDigdagWorkerAutoscaler.Spec.Deployment,
		postgresql:       horizontalDigdagWorkerAutoscaler.Spec.Postgresql,
		scaleIntervalSec: 15,
		db:               db,
	}

	cron := cron.New()
	cron.AddFunc("*/15 * * * * *", scaler.scaleDigdagWorker)
	scaler.cron = cron
	scaler.cron.Start()

	return scaler, nil
}
