# digdag-worker-crd

![Publish](https://github.com/TrsNium/digdag-worker-crd/workflows/Publish/badge.svg)

`digdag-worker-crd` is made for `digdag` project.
digdag is a workflow engine. This project aims to make digdag worker scalable on kubernetes.

## Usage

```yaml
apiVersion: horizontalpodautoscalers.autoscaling.digdag-worker-crd/v1
kind: HorizontalDigdagWorkerAutoscaler
metadata:
  name: horizontaldigdagworkerautoscaler-sample
spec:
  scaleTargetDeployment: digdag-worker
  scaleTargetDeploymentNamespace: default
  digdagMaxTaskThreads: 3
  postgresqlHost: "example.com"
  postgresqlPort: 5432
  postgresqlDatabase: "digdag"
  postgresqlUser: "digdag"
  postgresqlPassword: "your password"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    run: digdag-worker
  name: digdag-worker
spec:
  progressDeadlineSeconds: 600
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      run: digdag-worker
      containers:
      - args:
        - -cx
        - digdag server --disable-scheduler --config /etc/config/digdag.properties  --max-task-threads 3
        command:
        - /bin/bash
        image: asia.gcr.io/wear-hairstyle-stg/digdag:latest
        imagePullPolicy: Always
        name: digdag-worker
        dnsPolicy: ClusterFirst
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext: {}
        volumes:
        - configMap:
            name: digdag-config
          name: digdag-config-volume
```

## Deteil
`HorizontalDigdagWorkerAutoscaler` looks at the postgresql task queue used by the digdag worker and adjusts the digdag worker's replicas.
If the task queue is empty, set replicas of digdag worker to 1.
Increase the replicas of the digdag worker under the following conditions.
`TotalQueudTask - digdagMaxTaskThreads * DigdagWorker Replicas > 0`


