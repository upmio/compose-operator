# Redis Sentinel Failed Source Label Cleanup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove stale routing labels from an unhealthy `spec.source` Pod while preserving Sentinel's existing responsibility for writing the elected source back to `RedisReplication.spec.source`.

**Architecture:** Keep `spec.source` as the desired source and the Sentinel unit-agent callback as the only failover writer. Add a pure source-health helper and a pure failed-source cleanup helper, then use them during resource reconciliation so replica and Sentinel Pods retain their last valid source address until Sentinel updates the spec.

**Tech Stack:** Go, controller-runtime fake client, Kubernetes core/v1 API, Testify, existing Ginkgo/envtest integration suite.

---

## File Map

- Modify: `controller/redisreplication/handler.go`
  - Determine whether `spec.source` is healthy for routing.
  - Remove only failed-source routing labels.
  - Preserve replica and Sentinel source-address labels during failover.
- Create: `controller/redisreplication/handler_labels_test.go`
  - Fast unit coverage that runs without starting envtest.
- Reference: `controller/redisreplication/status.go:159-244`
  - Unavailable nodes become `Role=None`, `Status=KO`.
- Reference: `controller/redisreplication/redisreplication_controller.go:119-151`
  - Resource reconciliation still runs while Sentinel takeover skips topology mutation.

### Task 1: Add Source Health and Cleanup Primitives

**Files:**
- Modify: `controller/redisreplication/handler.go:295-380`
- Create: `controller/redisreplication/handler_labels_test.go`

- [ ] **Step 1: Write the failing source-health test**

Create `handler_labels_test.go` with these imports and fixtures:

~~~go
package redisreplication

import (
	"testing"

	"github.com/stretchr/testify/require"
	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newRoutingInstance() *composev1alpha1.RedisReplication {
	return &composev1alpha1.RedisReplication{
		ObjectMeta: metav1.ObjectMeta{Name: "redis", Namespace: "default"},
		Spec: composev1alpha1.RedisReplicationSpec{
			Source: &composev1alpha1.RedisNode{
				CommonNode: composev1alpha1.CommonNode{
					Name: "redis-0",
					Host: "redis-0.redis-headless-svc.default",
					Port: 6379,
				},
				AnnounceHost: "redis-0-svc.default.svc",
				AnnouncePort: 6379,
			},
		},
		Status: composev1alpha1.RedisReplicationStatus{
			Topology: composev1alpha1.RedisReplicationTopology{
				"redis-0": {
					Role:         composev1alpha1.RedisReplicationNodeRoleSource,
					Status:       composev1alpha1.NodeStatusOK,
					AnnounceHost: "redis-0-svc.default.svc",
					AnnouncePort: 6379,
				},
			},
		},
	}
}

func TestSourceRoutingLabels(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(*composev1alpha1.RedisReplication)
		wantHost string
		wantPort string
		wantOK   bool
	}{
		{"healthy", func(*composev1alpha1.RedisReplication) {}, "redis-0-svc.default.svc", "6379", true},
		{"missing", func(i *composev1alpha1.RedisReplication) {
			delete(i.Status.Topology, "redis-0")
		}, "", "", false},
		{"unreachable", func(i *composev1alpha1.RedisReplication) {
			i.Status.Topology["redis-0"].Status = composev1alpha1.NodeStatusKO
		}, "", "", false},
		{"demoted", func(i *composev1alpha1.RedisReplication) {
			i.Status.Topology["redis-0"].Role = composev1alpha1.RedisReplicationNodeRoleReplica
		}, "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := newRoutingInstance()
			tt.mutate(instance)
			host, port, ok := sourceRoutingLabels(instance)
			require.Equal(t, tt.wantHost, host)
			require.Equal(t, tt.wantPort, port)
			require.Equal(t, tt.wantOK, ok)
		})
	}
}
~~~

- [ ] **Step 2: Write the failing exact-cleanup and idempotency test**

~~~go
func TestClearFailedSourcePodLabels(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
		defaultKey: "redis", readOnlyKey: "false",
		sourceHostKey: "redis-0-svc.default.svc", sourcePortKey: "6379",
		"keep-me": "true",
	}}}

	changed, _ := clearFailedSourcePodLabels(pod, "redis")
	require.True(t, changed)
	require.Equal(t, "redis", pod.Labels[defaultKey])
	require.Equal(t, "true", pod.Labels["keep-me"])
	require.NotContains(t, pod.Labels, readOnlyKey)
	require.NotContains(t, pod.Labels, sourceHostKey)
	require.NotContains(t, pod.Labels, sourcePortKey)

	changed, _ = clearFailedSourcePodLabels(pod, "redis")
	require.False(t, changed)
}
~~~

- [ ] **Step 3: Run the focused tests and observe the expected failure**

Run:

~~~bash
go test ./controller/redisreplication -run 'Test(SourceRoutingLabels|ClearFailedSourcePodLabels)' -count=1
~~~

Expected: compilation fails with undefined `sourceRoutingLabels` and `clearFailedSourcePodLabels`.

- [ ] **Step 4: Implement the pure helpers in `handler.go`**

~~~go
func sourceRoutingLabels(instance *composev1alpha1.RedisReplication) (string, string, bool) {
	if instance == nil || instance.Spec.Source == nil {
		return "", "", false
	}
	sourceStatus, found := instance.Status.Topology[instance.Spec.Source.Name]
	if !found || sourceStatus == nil ||
		sourceStatus.Role != composev1alpha1.RedisReplicationNodeRoleSource ||
		sourceStatus.Status != composev1alpha1.NodeStatusOK {
		return "", "", false
	}
	return sourceStatus.AnnounceHost, strconv.Itoa(sourceStatus.AnnouncePort), true
}

func clearFailedSourcePodLabels(pod *corev1.Pod, instanceName string) (bool, string) {
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	changed := false
	changedLabels := make([]string, 0, 4)
	if pod.Labels[defaultKey] != instanceName {
		pod.Labels[defaultKey] = instanceName
		changed = true
		changedLabels = append(changedLabels, defaultKey)
	}
	for _, key := range []string{readOnlyKey, sourceHostKey, sourcePortKey} {
		if _, found := pod.Labels[key]; found {
			delete(pod.Labels, key)
			changed = true
			changedLabels = append(changedLabels, key)
		}
	}
	if !changed {
		return false, ""
	}
	return true, fmt.Sprintf("update labels '%s' successfully",
		strings.Join(changedLabels, ", "))
}
~~~

- [ ] **Step 5: Run the focused tests**

Run the Step 3 command again.

Expected: `ok github.com/upmio/compose-operator/controller/redisreplication`.

- [ ] **Step 6: Commit Task 1**

~~~bash
git add controller/redisreplication/handler.go controller/redisreplication/handler_labels_test.go
git commit -m "fix(redis-replication): define failed source label cleanup" \
  -m "Classify source routing health from observed topology and remove only stale routing labels while preserving the RedisReplication ownership label."
~~~

### Task 2: Exclude an Unhealthy Source Pod from Write Routing

**Files:**
- Modify: `controller/redisreplication/handler.go:65-79,295-380`
- Test: `controller/redisreplication/handler_labels_test.go`

- [ ] **Step 1: Add fake-client test utilities**

Extend the test import block with:

~~~go
"context"

"k8s.io/apimachinery/pkg/runtime"
"k8s.io/apimachinery/pkg/types"
"k8s.io/client-go/tools/record"
"sigs.k8s.io/controller-runtime/pkg/client"
"sigs.k8s.io/controller-runtime/pkg/client/fake"
~~~

Then append:

~~~go
func newLabelTestReconciler(t *testing.T, objects ...client.Object) *ReconcileRedisReplication {
	t.Helper()
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	return &ReconcileRedisReplication{
		client: fake.NewClientBuilder().WithScheme(scheme).
			WithObjects(objects...).Build(),
		recorder: record.NewFakeRecorder(20),
	}
}

func getTestPod(t *testing.T, r *ReconcileRedisReplication, name string) *corev1.Pod {
	t.Helper()
	pod := &corev1.Pod{}
	require.NoError(t, r.client.Get(context.Background(), types.NamespacedName{
		Name: name, Namespace: "default",
	}, pod))
	return pod
}
~~~

- [ ] **Step 2: Write the failing source reconciliation test**

~~~go
func TestEnsureSourcePodLabelsClearsUnhealthySource(t *testing.T) {
	instance := newRoutingInstance()
	instance.Status.Topology["redis-0"].Status = composev1alpha1.NodeStatusKO
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name: "redis-0", Namespace: "default",
		Labels: map[string]string{
			defaultKey: "redis", readOnlyKey: "false",
			sourceHostKey: "redis-0-svc.default.svc", sourcePortKey: "6379",
		},
	}}
	r := newLabelTestReconciler(t, pod)

	err := r.ensureSourcePodLabels(&syncContext{
		ctx: context.Background(), instance: instance,
	})
	require.NoError(t, err)

	updated := getTestPod(t, r, "redis-0")
	require.Equal(t, "redis", updated.Labels[defaultKey])
	require.NotContains(t, updated.Labels, readOnlyKey)
	require.NotContains(t, updated.Labels, sourceHostKey)
	require.NotContains(t, updated.Labels, sourcePortKey)
}
~~~

- [ ] **Step 3: Run the test and observe the expected failure**

~~~bash
go test ./controller/redisreplication -run TestEnsureSourcePodLabelsClearsUnhealthySource -count=1
~~~

Expected: compilation fails because `ensureSourcePodLabels` does not exist.

- [ ] **Step 4: Implement dedicated source Pod reconciliation**

Add:

~~~go
func (r *ReconcileRedisReplication) ensureSourcePodLabels(syncCtx *syncContext) error {
	instance := syncCtx.instance
	pod := &corev1.Pod{}
	if err := r.client.Get(syncCtx.ctx, types.NamespacedName{
		Name: instance.Spec.Source.Name, Namespace: instance.Namespace,
	}, pod); err != nil {
		return fmt.Errorf("failed to fetch pod [%s]: %v",
			instance.Spec.Source.Name, err)
	}

	host, port, healthy := sourceRoutingLabels(instance)
	var changed bool
	var message string
	if healthy {
		changed, message = r.setLabelsOnPod(
			pod, instance.Name, "false", host, port, true)
	} else {
		changed, message = clearFailedSourcePodLabels(pod, instance.Name)
	}
	if !changed {
		return nil
	}
	if err := r.client.Update(syncCtx.ctx, pod); err != nil {
		return fmt.Errorf("failed to update pod [%s]: %v", pod.Name, err)
	}
	r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced,
		"pod [%s] %s", pod.Name, message)
	return nil
}
~~~

Change `handleResources` to call `ensureSourcePodLabels(syncCtx)` for `spec.source`.

Change the function signature and guard only the source-host and source-port assignments:

~~~go
func (r *ReconcileRedisReplication) setLabelsOnPod(
	pod *corev1.Pod,
	instanceName, isReadOnly, hostLabelValue, portLabelValue string,
	updateSourceLabels bool,
) (bool, string) {
	var needsUpdate bool
	var updatedLabels []string

	if readOnlyValue, ok := pod.Labels[readOnlyKey]; !ok || readOnlyValue != isReadOnly {
		pod.Labels[readOnlyKey] = isReadOnly
		needsUpdate = true
		updatedLabels = append(updatedLabels, readOnlyKey)
	}
	if instanceValue, ok := pod.Labels[defaultKey]; !ok || instanceValue != instanceName {
		pod.Labels[defaultKey] = instanceName
		needsUpdate = true
		updatedLabels = append(updatedLabels, defaultKey)
	}

	if updateSourceLabels {
		if currentHost, ok := pod.Labels[sourceHostKey]; !ok || currentHost != hostLabelValue {
			pod.Labels[sourceHostKey] = hostLabelValue
			needsUpdate = true
			updatedLabels = append(updatedLabels, sourceHostKey)
		}
		if currentPort, ok := pod.Labels[sourcePortKey]; !ok || currentPort != portLabelValue {
			pod.Labels[sourcePortKey] = portLabelValue
			needsUpdate = true
			updatedLabels = append(updatedLabels, sourcePortKey)
		}
	}

	if !needsUpdate {
		return false, ""
	}
	return true, fmt.Sprintf("update labels '%s' successfully",
		strings.Join(updatedLabels, ", "))
}
~~~

To keep Task 2 scoped to source cleanup, update the existing replica call to pass `true` temporarily:

~~~go
needsUpdate, eventMessage = r.setLabelsOnPod(
	foundPod, instance.Name, isReadOnly,
	hostLabelValue, portLabelValue, true,
)
~~~

Task 3 replaces this temporary value with observed source health.

- [ ] **Step 5: Add the healthy-source companion test**

Append the complete healthy-source test:

~~~go
func TestEnsureSourcePodLabelsSetsHealthySource(t *testing.T) {
	instance := newRoutingInstance()
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name: "redis-0", Namespace: "default",
		Labels: map[string]string{},
	}}
	r := newLabelTestReconciler(t, pod)

	err := r.ensureSourcePodLabels(&syncContext{
		ctx: context.Background(), instance: instance,
	})
	require.NoError(t, err)

	updated := getTestPod(t, r, "redis-0")
	require.Equal(t, "redis", updated.Labels[defaultKey])
	require.Equal(t, "false", updated.Labels[readOnlyKey])
	require.Equal(t, "redis-0-svc.default.svc", updated.Labels[sourceHostKey])
	require.Equal(t, "6379", updated.Labels[sourcePortKey])
}
~~~

- [ ] **Step 6: Run source tests**

~~~bash
go test ./controller/redisreplication -run 'TestEnsureSourcePodLabels' -count=1
~~~

Expected: healthy and unhealthy cases pass.

- [ ] **Step 7: Commit Task 2**

~~~bash
git add controller/redisreplication/handler.go controller/redisreplication/handler_labels_test.go
git commit -m "fix(redis-replication): exclude unhealthy source from writes" \
  -m "Remove stale routing labels when the spec source is unavailable or demoted so the read-write Service fails closed until Sentinel writes the elected source back to the spec."
~~~

### Task 3: Preserve Replica and Sentinel Bootstrap Labels

**Files:**
- Modify: `controller/redisreplication/handler.go:295-341,423-478`
- Test: `controller/redisreplication/handler_labels_test.go`

- [ ] **Step 1: Write the failing replica preservation test**

~~~go
func TestEnsurePodLabelsPreservesReplicaSourceWhenSourceUnhealthy(t *testing.T) {
	instance := newRoutingInstance()
	instance.Status.Topology["redis-0"].Status = composev1alpha1.NodeStatusKO
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name: "redis-1", Namespace: "default",
		Labels: map[string]string{
			defaultKey: "redis", readOnlyKey: "true",
			sourceHostKey: "redis-0-svc.default.svc", sourcePortKey: "6379",
		},
	}}
	r := newLabelTestReconciler(t, pod)

	err := r.ensurePodLabels(&syncContext{
		ctx: context.Background(), instance: instance,
	}, "redis-1", "true", false)
	require.NoError(t, err)

	updated := getTestPod(t, r, "redis-1")
	require.Equal(t, "redis-0-svc.default.svc", updated.Labels[sourceHostKey])
	require.Equal(t, "6379", updated.Labels[sourcePortKey])
}
~~~

- [ ] **Step 2: Write the failing Sentinel preservation test**

~~~go
func TestEnsureSentinelPodLabelsPreservesSourceWhenSourceUnhealthy(t *testing.T) {
	instance := newRoutingInstance()
	instance.Spec.Sentinel = []string{"sentinel-0"}
	instance.Status.Topology["redis-0"].Status = composev1alpha1.NodeStatusKO
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name: "sentinel-0", Namespace: "default",
		Labels: map[string]string{
			sourceHostKey: "redis-0-svc.default.svc", sourcePortKey: "6379",
		},
	}}
	r := newLabelTestReconciler(t, pod)

	err := r.ensureSentinelPodLabels(&syncContext{
		ctx: context.Background(), instance: instance,
	})
	require.NoError(t, err)

	updated := getTestPod(t, r, "sentinel-0")
	require.Equal(t, "redis-0-svc.default.svc", updated.Labels[sourceHostKey])
	require.Equal(t, "6379", updated.Labels[sourcePortKey])
}
~~~

- [ ] **Step 3: Run the two tests and observe the expected failure**

~~~bash
go test ./controller/redisreplication -run 'TestEnsure(PodLabelsPreserves|SentinelPodLabelsPreserves)' -count=1
~~~

Expected before implementation: source host and port are empty.

- [ ] **Step 4: Preserve replica labels**

Replace duplicated source-status calculation in `ensurePodLabels` with:

~~~go
host, port, updateSourceLabels := sourceRoutingLabels(instance)
~~~

Replace the temporary Task 2 call with:

~~~go
needsUpdate, eventMessage = r.setLabelsOnPod(
	foundPod, instance.Name, isReadOnly,
	host, port, updateSourceLabels,
)
~~~

Readonly and replication-name labels still reconcile when `updateSourceLabels` is false.

- [ ] **Step 5: Preserve Sentinel labels**

At the beginning of `ensureSentinelPodLabels`, use:

~~~go
hostLabelValue, portLabelValue, healthy := sourceRoutingLabels(instance)
if !healthy {
	return nil
}
~~~

The existing update loop then writes `hostLabelValue` and `portLabelValue` only in the healthy case.

- [ ] **Step 6: Run all focused label tests**

~~~bash
go test ./controller/redisreplication -run 'Test(SourceRoutingLabels|ClearFailedSourcePodLabels|EnsureSourcePodLabels|EnsurePodLabelsPreserves|EnsureSentinelPodLabelsPreserves)' -count=1
~~~

Expected: all tests pass without starting envtest.

- [ ] **Step 7: Commit Task 3**

~~~bash
git add controller/redisreplication/handler.go controller/redisreplication/handler_labels_test.go
git commit -m "fix(redis-replication): preserve failover bootstrap labels" \
  -m "Keep the last valid source address on replica and Sentinel Pods while Sentinel is electing and persisting a replacement source."
~~~

### Task 4: Verify the Complete Sentinel Writeback Transition

**Files:**
- Modify: `controller/redisreplication/handler_labels_test.go`
- Verify: `controller/redisreplication/handler.go`

- [ ] **Step 1: Write a transition test**

Add the complete transition test. It mutates `spec.source` explicitly, representing Sentinel writeback, and does not invoke Compose Operator source election:

~~~go
func TestSourceLabelsFollowSentinelSpecWriteback(t *testing.T) {
	instance := newRoutingInstance()
	instance.Status.Topology["redis-0"].Status = composev1alpha1.NodeStatusKO
	oldSource := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name: "redis-0", Namespace: "default",
		Labels: map[string]string{
			defaultKey: "redis", readOnlyKey: "false",
			sourceHostKey: "redis-0-svc.default.svc", sourcePortKey: "6379",
		},
	}}
	newSource := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name: "redis-1", Namespace: "default",
		Labels: map[string]string{defaultKey: "redis", readOnlyKey: "true"},
	}}
	r := newLabelTestReconciler(t, oldSource, newSource)

	require.NoError(t, r.ensureSourcePodLabels(&syncContext{
		ctx: context.Background(), instance: instance,
	}))
	cleaned := getTestPod(t, r, "redis-0")
	require.NotContains(t, cleaned.Labels, readOnlyKey)
	require.NotContains(t, cleaned.Labels, sourceHostKey)
	require.NotContains(t, cleaned.Labels, sourcePortKey)

	instance.Spec.Source = &composev1alpha1.RedisNode{
		CommonNode: composev1alpha1.CommonNode{
			Name: "redis-1", Host: "redis-1.redis-headless-svc.default", Port: 6379,
		},
		AnnounceHost: "redis-1-svc.default.svc",
		AnnouncePort: 6379,
	}
	instance.Status.Topology = composev1alpha1.RedisReplicationTopology{
		"redis-1": {
			Role: composev1alpha1.RedisReplicationNodeRoleSource,
			Status: composev1alpha1.NodeStatusOK,
			AnnounceHost: "redis-1-svc.default.svc",
			AnnouncePort: 6379,
		},
	}

	require.NoError(t, r.ensureSourcePodLabels(&syncContext{
		ctx: context.Background(), instance: instance,
	}))
	updated := getTestPod(t, r, "redis-1")
	require.Equal(t, "false", updated.Labels[readOnlyKey])
	require.Equal(t, "redis-1-svc.default.svc", updated.Labels[sourceHostKey])
	require.Equal(t, "6379", updated.Labels[sourcePortKey])
}
~~~

- [ ] **Step 2: Run the transition test**

~~~bash
go test ./controller/redisreplication -run TestSourceLabelsFollowSentinelSpecWriteback -count=1
~~~

Expected: PASS.

- [ ] **Step 3: Format and run static verification**

~~~bash
gofmt -w controller/redisreplication/handler.go controller/redisreplication/handler_labels_test.go
go vet ./controller/redisreplication ./pkg/redisutil
go build ./controller/redisreplication ./pkg/redisutil
git diff --check
~~~

Expected: every command exits zero; `git diff --check` prints nothing.

- [ ] **Step 4: Run focused and Redis utility tests**

~~~bash
go test ./controller/redisreplication -run 'Test(SourceRoutingLabels|ClearFailedSourcePodLabels|EnsureSourcePodLabels|EnsurePodLabelsPreserves|EnsureSentinelPodLabelsPreserves|SourceLabelsFollowSentinelSpecWriteback)' -count=1
go test ./pkg/redisutil -count=1
~~~

Expected: both commands report `ok`.

- [ ] **Step 5: Run the repository test target**

~~~bash
make test
~~~

Expected: envtest assets are installed under `bin/k8s`, non-e2e tests pass, and `cover.out` is generated. If unrelated pre-existing tests fail, record the exact failure without weakening focused coverage.

- [ ] **Step 6: Confirm scope**

~~~bash
git status --short
git diff --stat
git diff -- controller/redisreplication/handler.go controller/redisreplication/handler_labels_test.go
~~~

Expected: implementation changes are limited to the two planned files. Do not stage the pre-existing `webhook/redisreplication/redisreplication_webhook.go` modification or any `.DS_Store` files.

- [ ] **Step 7: Commit transition coverage**

~~~bash
git add controller/redisreplication/handler_labels_test.go
git commit -m "test(redis-replication): cover Sentinel source label transition" \
  -m "Verify fail-closed cleanup of the old source and restoration of write routing only after Sentinel persists the elected source in the spec."
~~~

- [ ] **Step 8: Review final history and worktree state**

~~~bash
git log -4 --format=fuller
git status --short
~~~

Expected: scoped Conventional Commit messages include explanatory bodies; only pre-existing user changes and untracked `.DS_Store` files remain outside the implementation commits.
