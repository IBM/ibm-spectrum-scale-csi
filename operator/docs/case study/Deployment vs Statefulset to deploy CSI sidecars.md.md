# Using deployment over statefulset for CSI Sidecars

## Table of Contents

- [Overview](#[-overview)
- [Issue with current implementation](#[-issue-with-current-implementation)
- [Root Cause Analysis](#[-root-cause-analysis)
- [Proposed Solution](#[-proposed-solution)
- [Implementation](#[-implementation)
- [Upgrading](#[-upgrading)
- [Reference](#[-reference])

---

## Overview

As per the [documentation](https://kubernetes-csi.github.io/docs/deploying.html), the controller component can be deployed as a Deployment or StatefulSet on any node in the cluster. Currently the CSI operator deploys the sidecars controllers as Statefulset.

Here's the list of controllers and resources used to deploy them.

|Controller|Resource Type|Replica|
| --- | --- | ---|
|Attacher|Statefulset| One|
|Provisioner|Statefulset| One |
|Snapshotter|Statefulset| One |
|Resizer|Statefulset| One |

## Issue with current implementation

If Attacher has one replica and is deployed as statefulset, and if that node is down. Attaching volumes fails for whole cluster.
Same for provisioning, resizing and snapshots.

Once node is up, statefulset pod comes to running state and pending volumes gets attached. Same for provisioning(pvc get bound), resizing and snapshots.

Reference: [PVC attached to a pod doesn't migrate across nodes when Kubelet Service is stopped](https://github.com/IBM/ibm-spectrum-scale-csi/issues/563)

## Root Cause Analysis

As per the [documentation](https://kubernetes.io/docs/tasks/run-application/force-delete-stateful-set-pod/#delete-pods) if the kubelet services stops responding in one the nodes or the nodes becomes unreachable, the Statefulset pods goes into terminating state and doesn't re-schedule to another node.

This behavior of the Statefulset is as per it's design. This ensures only one instance of a pod with given identity is running on the cluster.

## Some Possible scenarios to deploy a CSI Sidecar

| Resource | Replica | LeaderElection |
|---|---|---|
|Statefulset| One | False |
|Statefulset| Many | True |
|Deployment| One | True |
|Deployment| Many | True |

### Sidecar deployed as Statefulset with one replica

- **What happens during node failure**

  - Statefulset pods goes to terminating state upon node failure.
  - Pods stay in `terminating` state unless forcefully deleted using `--force` flag.
  - This behavior of Statefulset is as per its design.
  - Pods need to be forcefully deleted to re-schedule them on healthy node.
  - Forceful deletion is not preferred. Refer [Statefulset-Considerations]("#https://kubernetes.io/docs/tasks/run-application/force-delete-stateful-set-pod/#statefulset-considerations")
  - Pods do not re-schedule even if liveness probe is used.

This should not be recommended because if a node with sidecar pod goes down, the service stops handing requests.
Example: If attacher pod is scheduled in a node and the node crashes. Volume attachment will no longer work.

### Sidecar deployed as Deployment with one replica and leader election enabled

- **What happens during node failure**

  - Deployment pod goes to terminating state upon node failure.
  - ReplicaSet created by deployment creates a new pod in available healthy node.
  - New pod is elected as leader after lease timeout.
  - New pod starts accepting the requests.
  - Terminating pod stays in terminating state until node is up, once node is up the pod gets deleted.

This is a safer implementation. Leader-election prevents the chances of mis-handling requests, in case more then one instances of sidecar controller pod gets created by replica set during upgrade.

### Sidecar deployed as Deployment with multiple replicas and leader election enabled

- **What happens during node failure**

  - Deployment pod goes to terminating state upon node failure.
  - ReplicaSet created by deployment creates a new pod in available healthy node.
  - Any one out of the available running pods is elected as leader after lease timeout.
  - Lease holder pod starts accepting the requests.
  - Terminating pod stays in terminating state until node is up, once node is up the pod gets deleted.

In this implementation, number of running pods always stays as per replica count. However, this overall consumes more resource especially during upgrade. Also, if pod-anti-affinity is set to ensure pods are scheduled in unique nodes, then new pods during upgrade gets stuck in `pending` state sometimes. More details under section [Ensuring only one sidecar pod per worker node](#Ensuring-only-one-sidecar-pod-per-worker-node)

### Sidecar deployed as Statefulset with multiple replicas and leader election enabled

- **What happens during node failure**

  - Statefulset pods goes to terminating state upon node failure.
  - Pods stay in `terminating` state unless forcefully deleted using `--force` flag.
  - Any one out of the available runnings pods is elected as leader after lease timeout.
  - Lease holder pod starts accepting the requests.
  - Terminating pod stays in terminating state until node is up, once node is up the pod comes to running state.

Since Statefulset does not try to re-create new pod upon node failure, resource consumption is fixed.
However, since replica count of healthy pods decreases and only the available running pods take part in leader election and serve request.

## Proposed Solution

Rather than waiting for the node to recover, controller must ensure high availability by having replicas.
Leader election should be enabled to ensure only one replica/instance is serving the requests at a given time.
Must ensure replicas run on separate nodes, so that other replicas remain healthy when a node goes down.

### Solution 1

Using Statefulsets with replicas and leader election

-Replica: 2
-Leader Election: Enabled
-Pod Anti Affinity: Enabled
-Liveness Probe: Enabled
-Node Toleration: Optional

- **Reasons to deploy using Statefulset**

  - Lesser code change.
  - No challenges during upgrade.

- **Known issues**

  - This approach is not very common as recommended way to deploy a stateless application with replication is using Deployment.

### Solution 2

Since Statefulset and Deployment both are recommended resources to deploy the sidecar controllers. Statefulsets can be replaced with Deployments.
Using Statefulsets with replicas and leader election

- Replica: 2
- Leader Election: Enabled
- Pod Anti Affinity: Enabled
- Liveness Probe: Enabled
- Node Toleration: Optional

- **Reasons to deploy the controller using Deployment**

  ReplicaSet created by deployment will re-schedule the pod to another node once existing pods go to terminating state. This will ensure number of running pods is always as per the replica count.

- **Known issues**

  - Upgrade issue while using podAntiAffinity. Requires additional node to be available to complete the upgrade. Check section [Ensuring only one sidecar pod per worker node](#Ensuring-only-one-sidecar-pod-per-worker-node)

  - More resource utilization.

### Estimate time taken for the servcie to fully recover after node failure

Time taken using both the solutions is same as given below:

|Process| Time Taken |
|---|---|
|Time taken to update node state as `Not Ready`| ~30-40 seconds|
|Time taken to update pod state as `terminating`|300 seconds(default)|
|Time taken for current lease to expire| ~120 seconds|
|Total|~480 seconds(8 minutes)|

> Note: It takes around 30-60 seconds of time(depends on cluster) for the node controller to update the state of the node as *NotReady*  when kubelet is not reachable. Durning this time period the controllers shows connection refused error in logs.
>
> Note: Since the default pod eviction timeout value is 5 minutes, it takes another exact 5 mins for the pod to go into terminating state.
>
> Note: It takes another ~120 seconds for the terminating pod to release lease.

#### Handling pod eviction timeout

To reduce the overall time taken for the pod to go to terminating state and release the lease which is around 8 minutes, pod-eviction time can be decreased.
  
- Tolerations can be added under pod template in deployment.

    **Example:**

    ```yaml
      tolerations:
      - effect: NoExecute
        key: node.kubernetes.io/unreachable
        operator: Exists
        tolerationSeconds: 10
      - effect: NoExecute
        key: node.kubernetes.io/not-ready
        operator: Exists
        tolerationSeconds: 10
    ```

    **Explanation:**

    Here tolerations are set for two conditions, first when node is unreachable and second when node is in node ready state.
    If the two conditions satisfies, based on `tolerationSeconds`, the pods with stay bound to the node for 10 seconds and then evicted.
    This will reduce pod eviction time default 5 mins to 10 seconds.

#### Ensuring only one sidecar pod per worker node

  In case if a node goes down the sidecar pod terminates. If the node was lease holder, lease will be released after timeout.
  Other replica pods should be available and in running state to get hold of the lease. For this it should be ensured the each replica pods run on separate node. `PodAntiAffinity` can be used for this.

- **Using podAntiAffinity**

  ```yaml
     affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - csi-sidecar-controller
            topologyKey: kubernetes.io/hostname
  ```

- **Issue**

    Pod Anti Affinity affects the upgrade. New pods stays in pending state some times.

- **Reason**

    Old pods with matching labels are already present on the node, so new pods don't get scheduled.

- **Possible Solution**

    Updating the update strategy.

    ```yaml
    strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 50%
    type: RollingUpdate
    ```

### Implementation

Solution1: Existing statefulset implementation should be updated to use leader election and replicas.
Solution 2: The CSI Operator can be configured to deploy the sidecar controller pods using Deployment resource.

### Upgrading

Solution 2: Durning upgrading the CSI operator has to deploy the pods using Deployment resource and once pods are up and running, remove the Statefulset resource from the cluster.

> Note: Continuously checking the state of the sidecar controller pod, might require design discussion. For the current POC, the operator deploys the pods using deployment and then deletes the statefulsets.

### What if node is Ready, but DisabledScheduling

  Make node Ready but Unhealthy using `kubectl cordon <node>`

  Existing pod stays in running state, new pods that tries to schedule the node fails scheduling.

  Use `kubectl uncordon <node>` to enable Scheduling.

### Can we control lease timeout?

Given below args can be passed to `external-provisioner`,`external-attacher`,`external-resizer` & `external-snapshotter` to reduce lease time by some extent but this is not very effective. Needs more analysis.

- `--leader-election-lease-duration` <duration>: Duration, in seconds, that non-leader candidates will wait to force acquire leadership. Defaults to 15 seconds.

- `--leader-election-renew-deadline` <duration>: Duration, in seconds, that the acting leader will retry refreshing leadership before giving up. Defaults to 10 seconds.

- `--leader-election-retry-period` <duration>: Duration, in seconds, the LeaderElector clients should wait between tries of actions. Defaults to 5 seconds.

### Using single CSI controller pod with sidecar containers

**Pros:**

- This is the suggested way of deployment as per CSI design proposal. Refer [Design Proposal](https://github.com/kubernetes/design-proposals-archive/blob/main/storage/container-storage-interface.md)

- Many other CSI drivers have implementations in similar way. [CSI Drivers](https://kubernetes-csi.github.io/docs/drivers.html)

**Cons:**

- During upgrade or patch update, if any one sidecar is updated, then the pod will restart thus reloading other sidecars too.
- In case if liveness probe /healthz/leader-election of one container fails. This will restart the pod with all the containers. Though chances of leader-election liveness probe to fail are very low.

---

## Reference

- [Force Delete StatefulSet Pods](https://kubernetes.io/docs/tasks/run-application/force-delete-stateful-set-pod/#delete-pods)
- [PVC atached to a pod doesn't migrate across nodes when Kubelet Service is stopped](https://github.com/IBM/ibm-spectrum-scale-csi/issues/563)
- [Deploying CSI Driver on Kubernetes](https://kubernetes-csi.github.io/docs/deploying.html)
