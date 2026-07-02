package redisreplication

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

func TestClearFailedSourcePodLabels(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
		defaultKey:    "redis",
		readOnlyKey:   "false",
		sourceHostKey: "redis-0-svc.default.svc",
		sourcePortKey: "6379",
		"keep-me":     "true",
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

func newLabelTestReconciler(t *testing.T, objects ...client.Object) *ReconcileRedisReplication {
	t.Helper()
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	return &ReconcileRedisReplication{
		client:   fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build(),
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
			Role:         composev1alpha1.RedisReplicationNodeRoleSource,
			Status:       composev1alpha1.NodeStatusOK,
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
