# kube-globalreserve

## What is the kube-globalreserve

kube-globalreserve is a glogbal resource reserve tool that can work with Kubernetes default scheduler and cutomer scheduler.

Kubernetes allows to run separate custom scheduler along with the default scheduler. The default scheduler and custom scheduler covers respective pods (by spec.schedulerName) exclusively. However, co-existence of multiple schedulers might have more problems than it sounds like. You may see tricky issues when pods get scheduled onto the same node by multiple schedulers.

kube-globalreserve creates a global resource cache. And all schedulers reserve pod's resources globally before binding. This can clearly reduce the pod binding conflict effection, especially for Batch Job scheduler.

## Who need kube-globalreserve

If your customized scheduler dispatches pods one by one. You can reference [Multi Scheduling Profiles](https://github.com/kubernetes/enhancements/blob/master/keps/sig-scheduling/20200114-multi-scheduling-profiles.md) or [Scheduling Framework](https://github.com/kubernetes/enhancements/blob/master/keps/sig-scheduling/20180409-scheduling-framework.md)

If your cusomized scheduler dispatches pods by groups, you can refer this project.

## How does kube-globalreserve work

kube-globalreserve creates a global resource pool, every scheduler must reserve pods' resources beforing binding to nodes. If reserving failed, scheduler can recalculate.

### How to choose the correct API

kube-globalreserve supplies 2 kinds of interfaces:

1. Golang API
2. REST API

If default scheduler's performance is more important than your scheduler. The default scheduler uses Golang API and your scheduler uses REST API. The global resource pool stays in default scheduler's process.

If your scheduler's speed is more critical, your scheduler can use Golang API and default scheduler uses REST API. The global resource pool stays in your scheduler's process. (This way is still under development)

## Getting Started

### Prerequisites

Golang >= 1.12
Kubernetes >= 1.17.2

### Build

```bash
make
```

### Example

This example creates a scheduler with name `globalreserve-scheduler` to take the place of default scheduler and creates a deployment for testing the Golang API.

```bash
kubectl create namespace globalreserve-test
kubectl apply -f deployment/globalreserve-account.yaml
kubectl apply -f deployment/globalreserve-deployment.yaml

#following resources must specify schedulerName: globalreserve-scheduler
kubectl apply -f deployment/test.yaml
```

kube-globalreserve REST Server Port is "23456". Your scheduler can *POST* [PodsReserveRequest](./pkg/reserve/utils.go#L40) to `http://<hostname>:23456/reserve` for resource reservation before binding.

kube-globalreserve log can show reserve details.

### Replace Default Scheduler

You can using the build output `./bin/kube-globalreserve-scheduler` to replace the default scheduler binary and restart default scheduler.

### Configuration

(Under development)

### References

If you have any further question, please connect with [xq2005](https://github.com/xq2005).
