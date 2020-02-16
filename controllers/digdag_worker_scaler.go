package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-logr/logr"
	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
	"math"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	horizontalpodautoscalersautoscalingv1 "digdag-worker-crd/api/v1"
)

type DigdagWorkerScaler struct {
	client                client.Client
	log                   logr.Logger
	recorder              record.EventRecorder
	namespace             string
	postgresqlHost        string
	postgresqlPort        string
	postgresqlDatabase    string
	postgresqlUser        string
	postgresqlPassword    string
	scaleTargetDeployment string
	scaleIntervalSec      int32
	maxTaskThreads        int32
	cron                  *cron.Cron
	db                    *sql.DB
}

func (r *DigdagWorkerScaler) Equal(horizontalDigdagWorkerAutoscaler horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscaler) bool {
	spec := horizontalDigdagWorkerAutoscaler.Spec
	objectMeta := horizontalDigdagWorkerAutoscaler.ObjectMeta
	return r.namespace != objectMeta.Namespace ||
		r.postgresqlHost != spec.PostgresqlHost ||
		r.postgresqlPort != spec.PostgresqlPort ||
		r.postgresqlDatabase != spec.PostgresqlDatabase ||
		r.postgresqlUser != spec.PostgresqlUser ||
		r.postgresqlPassword != spec.PostgresqlPassword ||
		r.scaleTargetDeployment != spec.ScaleTargetDeployment ||
		r.scaleIntervalSec != 30 ||
		r.maxTaskThreads != spec.DigdagWorkerMaxTaskThreads
}

func (r *DigdagWorkerScaler) Update(horizontalDigdagWorkerAutoscaler horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscaler) error {
	r.GC()

	spec := horizontalDigdagWorkerAutoscaler.Spec
	objectMeta := horizontalDigdagWorkerAutoscaler.ObjectMeta
	db, err := createDB(spec.PostgresqlHost, spec.PostgresqlPort, spec.PostgresqlDatabase, spec.PostgresqlUser, spec.PostgresqlPassword)
	if err != nil {
		return err
	}

	scaler.cron = cron
	r.namespace = objectMeta.Namespace
	r.postgresqlHost = spec.PostgresqlHost
	r.postgresqlPort = spec.PostgresqlPort
	r.postgresqlDatabase = spec.PostgresqlDatabase
	r.postgresqlUser = spec.PostgresqlUser
	r.postgresqlPassword = spec.PostgresqlPassword
	r.scaleTargetDeployment = spec.ScaleTargetDeployment
	r.cronIntervalSec = 30
	r.maxTaskThreads = spec.DigdagWorkerMaxTaskThreads
	r.db = db

	cron := cron.New()
	cron.AddFunc()
	cron.Start()
	r.cron = cron
	return nil
}

func (r *DigdagWorkerScaler) getQueuedTaskNum(spec horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscalerSpec) (int32, error) {
	var count int32
	if err := r.db.QueryRow("select count(*) from tasks;").Scan(&name); err != nil {
		return nil, err
	}
	return count, nil
}

func (r *DigdagWorkerScaler) getRunningTaskNum(spec horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscalerSpec) (int32, error) {
	var count int32
	if err := r.db.QueryRow("select count(*) from tasks where task_type = 0 and state = 2;").Scan(&name); err != nil {
		return nil, err
	}
	return count, nil
}

func (r *DigdagWorkerScaler) scaleDigdagWorker() error {
	ctx := context.Background()
	// Get Deployment associated with HorizontalDigdagWorkerAutoscaler
	deployment := appsv1.Deployment{}
	err := r.client.Get(ctx, client.ObjectKey{Namespace: r.namespace, Name: r.scaleTargetDeployment}, &deployment)
	if errors.IsNotFound(err) {
		r.log.Info("Deployment associated with HorizontalDigdagWorkerAutoscaler was not found")
		return nil
	}
	if err != nil {
		r.log.Error(err, "failed to get Deployment for MyKind resource")
		return err
	}

	// Obtain digdag task queue info from HorizontalDigdagWorkerAutoscaler's configure
	queuedTaskNum, err := r.getQueuedTaskNum(horizontalDigdagWorkerAutoscaler.Spec)
	if err != nil {
		r.log.Error(err, "failed to get queuedTaskNum")
		return err
	}

	// If there are no queued tasks, set Replicas to 1.
	// If there are queued tasks, get the number of tasks that have not been executed and update the Replicas.
	if queuedTotalTaskNum == 0 {
		// Set replicas to 1 because there are no tasks to execute
		r.log.Info("Digdag is idling now")

		expectedReplicas := int32(1)
		deployment.Spec.Replicas = &expectedReplicas
		if err := r.client.Update(ctx, &deployment); err != nil {
			r.log.Error(err, "failed to Deployment update replica count")
			return err
		}

		return nil
	} else {
		runningTaskNum, err := r.getRunningTaskNum()
		if err != nil {
			r.log.Error(err, "failed to get planingTaskNum")
			return err
		}

		// Obtain replica of Deployment linked to HorizontalDigdagWorkerAutoscaler
		currentReplicas := *deployment.Spec.Replicas

		// Update the number of deployment pods according to the task queue
		digdagWorkerMaxTaskThreads := r.maxTaskThreads
		digdagTotalTaskThreads := currentReplicas * digdagWorkerMaxTaskThreads

		// NOTE
		// Tasks that are not running on any node will be in the running state,
		// So subtracting digdagTotalTaskThreads from all running tasks will give you the number of surplus tasks
		surplusTaskNum := runningTaskNum - digdagTotalTaskThreads
		if surplusTaskNum > 0 {
			additionalReplicas := int32(math.math.Ceil(surplusTaskNum / digdagWorkerMaxTaskThreads))
			newReplicas := currentReplicas + additionalReplicas

			deployment.Spec.Replicas = &newReplicas
			if err := r.client.Update(ctx, &deployment); err != nil {
				r.log.Error(err, "failed to Deployment update replica count")
				return err
			}

			return nil
		}
	}
}

func (r *DigdagWorkerScaler) GC() {
	r.postgresqlDb.Close()
	r.cron.Stop()
}

func createDB(host string, port string, database string, user string, password string) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=verify-full", host, port, database, user, password)
	return sql.Open("postgres", connStr)
}

func NewDigdagWorkerScaler(client client.Client, log logr.Logger, horizontalDigdagWorkerAutoscaler horizontalpodautoscalersautoscalingv1.HorizontalDigdagWorkerAutoscaler) (DigdagWorkerScaler, error) {
	spec := horizontalDigdagWorkerAutoscaler.Spec
	objectMeta := horizontalDigdagWorkerAutoscaler.ObjectMeta
	db, err := createDB(spec.PostgresqlHost, spec.PostgresqlPort, spec.PostgresqlDatabase, spec.PostgresqlUser, spec.PostgresqlPassword)
	if err != nil {
		return nil, err
	}

	scaler = DigdagWorkerScaler{
		client:                client,
		log:                   log,
		namespace:             objectMeta.Namespace,
		postgresqlHost:        spec.PostgresqlHost,
		postgresqlPort:        spec.PostgresqlPort,
		postgresqlDatabase:    spec.PostgresqlDatabase,
		postgresqlUser:        spec.PostgresqlUser,
		postgresqlPassword:    spec.PostgresqlPassword,
		scaleTargetDeployment: spec.ScaleTargetDeployment,
		scaleIntervalSec:      30,
		maxTaskThreads:        spec.DigdagWorkerMaxTaskThreads,
		db:                    db,
	}

	cron := cron.New()
	cron.AddFunc(fmt.Sprintf("*/%i * * * * *", scaler.scaleIntervalSec), scaler.scaleDigdagWorker())
	cron.Start()
	scaler.cron = cron
	return scaler, nil
}
