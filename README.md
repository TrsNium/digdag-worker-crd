# digdag-worker-crd

![Publish](https://github.com/TrsNium/digdag-worker-crd/workflows/Publish/badge.svg)

Digdag is a workflow engine.  
`digdag-worker-crd` is made for Digdag project.  
This project aims to make digdag worker scalable on kubernetes.

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
        image: yourImage
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
`HorizontalDigdagWorkerAutoscaler` looks at the postgresql task queue used by the Digdag worker and adjusts the Digdag worker's replicas.  

If the task queue is empty, set replicas of Digdag worker to 1.  
Increase the replicas of the Digdag worker under the following conditions.  
`TotalQueudTask - DigdagMaxTaskThreads * DigdagWorker Replicas > 0`  

Also Digdag worker scale-in will not be done emptying the task queue.
