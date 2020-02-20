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
  name: horizontaldigdagworkerautoscaler
spec:
  deployment:
    name: digdag-worker
    namespace: default
    maxTaskThreads: 3
  postgresql:
    host:
      value: <YOUR_POSTGRESQL_HOST>
    port:
      value: "5432"
    database:
      value: postgres
    user:
      value: postgres
    password:
      valueFromSecretKeyRef:
        name: postgres
        namespace: default
        key: password
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
        volumeMounts:
        - mountPath: /etc/config
          name: digdag-config-volume
      volumes:
      - configMap:
          name: digdag-config
        name: digdag-config-volume
```

## Detail
`HorizontalDigdagWorkerAutoscaler` looks at the postgresql task queue used by the Digdag worker and adjusts the Digdag worker's replicas.  

If the task queue is empty, set replicas of Digdag worker to 1.  
Increase the replicas of the Digdag worker under the following conditions.  
`TotalQueudTask - DigdagMaxTaskThreads * DigdagWorker Replicas > 0`  

Also Digdag worker scale-in will not be done emptying the task queue.
