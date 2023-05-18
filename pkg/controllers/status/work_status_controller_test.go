package status

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clusterv1alpha1 "github.com/karmada-io/karmada/pkg/apis/cluster/v1alpha1"
	workv1alpha1 "github.com/karmada-io/karmada/pkg/apis/work/v1alpha1"
	"github.com/karmada-io/karmada/pkg/sharedcli/ratelimiterflag"
	"github.com/karmada-io/karmada/pkg/util"
	"github.com/karmada-io/karmada/pkg/util/fedinformer/genericmanager"
	"github.com/karmada-io/karmada/pkg/util/gclient"
	"github.com/karmada-io/karmada/pkg/util/helper"
)

func newCluster(name string, clusterType string, clusterStatus metav1.ConditionStatus) *clusterv1alpha1.Cluster {
	return &clusterv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: clusterv1alpha1.ClusterSpec{},
		Status: clusterv1alpha1.ClusterStatus{
			Conditions: []metav1.Condition{
				{
					Type:   clusterType,
					Status: clusterStatus,
				},
			},
		},
	}
}

func TestWorkStatusController_Reconcile(t *testing.T) {
	tests := []struct {
		name      string
		c         WorkStatusController
		work      *workv1alpha1.Work
		ns        string
		expectRes controllerruntime.Result
		existErr  bool
	}{
		{
			name: "normal case",
			c: WorkStatusController{
				Client:                      fake.NewClientBuilder().WithScheme(gclient.NewSchema()).WithObjects(newCluster("cluster", clusterv1alpha1.ClusterConditionReady, metav1.ConditionTrue)).Build(),
				InformerManager:             genericmanager.GetInstance(),
				PredicateFunc:               helper.NewClusterPredicateOnAgent("test"),
				ClusterDynamicClientSetFunc: util.NewClusterDynamicClientSetForAgent,
				ClusterCacheSyncTimeout:     metav1.Duration{},
				RateLimiterOptions:          ratelimiterflag.Options{},
			},
			work: &workv1alpha1.Work{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "work",
					Namespace: "karmada-es-cluster",
				},
				Status: workv1alpha1.WorkStatus{
					Conditions: []metav1.Condition{
						{
							Type:   workv1alpha1.WorkApplied,
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			ns:        "karmada-es-cluster",
			expectRes: controllerruntime.Result{},
			existErr:  false,
		},
		{
			name: "work not exists",
			c: WorkStatusController{
				Client:                      fake.NewClientBuilder().WithScheme(gclient.NewSchema()).WithObjects(newCluster("cluster", clusterv1alpha1.ClusterConditionReady, metav1.ConditionTrue)).Build(),
				InformerManager:             genericmanager.GetInstance(),
				PredicateFunc:               helper.NewClusterPredicateOnAgent("test"),
				ClusterDynamicClientSetFunc: util.NewClusterDynamicClientSetForAgent,
				ClusterCacheSyncTimeout:     metav1.Duration{},
				RateLimiterOptions:          ratelimiterflag.Options{},
			},
			work: &workv1alpha1.Work{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "work-1",
					Namespace: "karmada-es-cluster",
				},
				Status: workv1alpha1.WorkStatus{
					Conditions: []metav1.Condition{
						{
							Type:   workv1alpha1.WorkApplied,
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			ns:        "karmada-es-cluster",
			expectRes: controllerruntime.Result{},
			existErr:  false,
		},
		{
			name: "work's DeletionTimestamp isn't zero",
			c: WorkStatusController{
				Client:                      fake.NewClientBuilder().WithScheme(gclient.NewSchema()).WithObjects(newCluster("cluster", clusterv1alpha1.ClusterConditionReady, metav1.ConditionTrue)).Build(),
				InformerManager:             genericmanager.GetInstance(),
				PredicateFunc:               helper.NewClusterPredicateOnAgent("test"),
				ClusterDynamicClientSetFunc: util.NewClusterDynamicClientSetForAgent,
				ClusterCacheSyncTimeout:     metav1.Duration{},
				RateLimiterOptions:          ratelimiterflag.Options{},
			},
			work: &workv1alpha1.Work{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "work",
					Namespace:         "karmada-es-cluster",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
				},
				Status: workv1alpha1.WorkStatus{
					Conditions: []metav1.Condition{
						{
							Type:   workv1alpha1.WorkApplied,
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			ns:        "karmada-es-cluster",
			expectRes: controllerruntime.Result{},
			existErr:  false,
		},
		{
			name: "work's status is not applied",
			c: WorkStatusController{
				Client:                      fake.NewClientBuilder().WithScheme(gclient.NewSchema()).WithObjects(newCluster("cluster", clusterv1alpha1.ClusterConditionReady, metav1.ConditionTrue)).Build(),
				InformerManager:             genericmanager.GetInstance(),
				PredicateFunc:               helper.NewClusterPredicateOnAgent("test"),
				ClusterDynamicClientSetFunc: util.NewClusterDynamicClientSetForAgent,
				ClusterCacheSyncTimeout:     metav1.Duration{},
				RateLimiterOptions:          ratelimiterflag.Options{},
			},
			work: &workv1alpha1.Work{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "work",
					Namespace: "karmada-es-cluster",
				},
				Status: workv1alpha1.WorkStatus{
					Conditions: []metav1.Condition{
						{
							Type:   workv1alpha1.WorkAvailable,
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			ns:        "karmada-es-cluster",
			expectRes: controllerruntime.Result{},
			existErr:  false,
		},
		{
			name: "failed to get cluster name",
			c: WorkStatusController{
				Client:                      fake.NewClientBuilder().WithScheme(gclient.NewSchema()).WithObjects(newCluster("cluster", clusterv1alpha1.ClusterConditionReady, metav1.ConditionTrue)).Build(),
				InformerManager:             genericmanager.GetInstance(),
				PredicateFunc:               helper.NewClusterPredicateOnAgent("test"),
				ClusterDynamicClientSetFunc: util.NewClusterDynamicClientSetForAgent,
				ClusterCacheSyncTimeout:     metav1.Duration{},
				RateLimiterOptions:          ratelimiterflag.Options{},
			},
			work: &workv1alpha1.Work{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "work",
					Namespace: "karmada-cluster",
				},
				Status: workv1alpha1.WorkStatus{
					Conditions: []metav1.Condition{
						{
							Type:   workv1alpha1.WorkApplied,
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			ns:        "karmada-cluster",
			expectRes: controllerruntime.Result{Requeue: true},
			existErr:  true,
		},
		{
			name: "failed to get cluster",
			c: WorkStatusController{
				Client:                      fake.NewClientBuilder().WithScheme(gclient.NewSchema()).WithObjects(newCluster("cluster1", clusterv1alpha1.ClusterConditionReady, metav1.ConditionTrue)).Build(),
				InformerManager:             genericmanager.GetInstance(),
				PredicateFunc:               helper.NewClusterPredicateOnAgent("test"),
				ClusterDynamicClientSetFunc: util.NewClusterDynamicClientSetForAgent,
				ClusterCacheSyncTimeout:     metav1.Duration{},
				RateLimiterOptions:          ratelimiterflag.Options{},
			},
			work: &workv1alpha1.Work{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "work",
					Namespace: "karmada-es-cluster",
				},
				Status: workv1alpha1.WorkStatus{
					Conditions: []metav1.Condition{
						{
							Type:   workv1alpha1.WorkApplied,
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			ns:        "karmada-es-cluster",
			expectRes: controllerruntime.Result{Requeue: true},
			existErr:  true,
		},
		{
			name: "cluster is not ready",
			c: WorkStatusController{
				Client:                      fake.NewClientBuilder().WithScheme(gclient.NewSchema()).WithObjects(newCluster("cluster", clusterv1alpha1.ClusterConditionReady, metav1.ConditionFalse)).Build(),
				InformerManager:             genericmanager.GetInstance(),
				PredicateFunc:               helper.NewClusterPredicateOnAgent("test"),
				ClusterDynamicClientSetFunc: util.NewClusterDynamicClientSetForAgent,
				ClusterCacheSyncTimeout:     metav1.Duration{},
				RateLimiterOptions:          ratelimiterflag.Options{},
			},
			work: &workv1alpha1.Work{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "work",
					Namespace: "karmada-es-cluster",
				},
				Status: workv1alpha1.WorkStatus{
					Conditions: []metav1.Condition{
						{
							Type:   workv1alpha1.WorkApplied,
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			ns:        "karmada-es-cluster",
			expectRes: controllerruntime.Result{Requeue: true},
			existErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := controllerruntime.Request{
				NamespacedName: types.NamespacedName{
					Name:      "work",
					Namespace: tt.ns,
				},
			}

			if err := tt.c.Client.Create(context.Background(), tt.work); err != nil {
				t.Fatalf("Failed to create cluster: %v", err)
			}

			res, err := tt.c.Reconcile(context.Background(), req)
			assert.Equal(t, tt.expectRes, res)
			if tt.existErr {
				assert.NotEmpty(t, err)
			} else {
				assert.Empty(t, err)
			}
		})
	}
}

func TestWorkStatusController_getEventHandler(t *testing.T) {
	opt := util.Options{
		Name:          "opt",
		KeyFunc:       nil,
		ReconcileFunc: nil,
	}

	c := WorkStatusController{
		Client:                      fake.NewClientBuilder().WithScheme(gclient.NewSchema()).WithObjects(newCluster("cluster", clusterv1alpha1.ClusterConditionReady, metav1.ConditionFalse)).Build(),
		InformerManager:             genericmanager.GetInstance(),
		PredicateFunc:               helper.NewClusterPredicateOnAgent("test"),
		ClusterDynamicClientSetFunc: util.NewClusterDynamicClientSetForAgent,
		ClusterCacheSyncTimeout:     metav1.Duration{},
		RateLimiterOptions:          ratelimiterflag.Options{},
		eventHandler:                nil,
		worker:                      util.NewAsyncWorker(opt),
	}

	eventHandler := c.getEventHandler()
	assert.NotEmpty(t, eventHandler)
}

func TestWorkStatusController_RunWorkQueue(t *testing.T) {
	c := WorkStatusController{
		Client:                      fake.NewClientBuilder().WithScheme(gclient.NewSchema()).WithObjects(newCluster("cluster", clusterv1alpha1.ClusterConditionReady, metav1.ConditionFalse)).Build(),
		InformerManager:             genericmanager.GetInstance(),
		PredicateFunc:               helper.NewClusterPredicateOnAgent("test"),
		ClusterDynamicClientSetFunc: util.NewClusterDynamicClientSetForAgent,
		ClusterCacheSyncTimeout:     metav1.Duration{},
		RateLimiterOptions:          ratelimiterflag.Options{},
		eventHandler:                nil,
	}

	c.RunWorkQueue()
}

func TestGenerateKey(t *testing.T) {
	tests := []struct {
		name     string
		resource *unstructured.Unstructured
		expect   string
		existErr bool
	}{
		{
			name: "normal case",
			resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "default",
						"labels": map[string]interface{}{
							workv1alpha1.WorkNamespaceLabel: "karmada-es-cluster",
						},
					},
				},
			},
			expect:   "cluster",
			existErr: false,
		},
		{
			name: "workNamespace is 0",
			resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "default",
						"labels": map[string]interface{}{
							"foo": "karmada-es-cluster",
						},
					},
				},
			},
			expect:   "",
			existErr: false,
		},
		{
			name: "failed to get cluster name",
			resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "default",
						"labels": map[string]interface{}{
							workv1alpha1.WorkNamespaceLabel: "karmada-cluster",
						},
					},
				},
			},
			expect:   "",
			existErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := getClusterNameFromLabel(tt.resource)
			assert.Equal(t, tt.expect, actual)
			if tt.existErr {
				assert.NotEmpty(t, err)
			} else {
				assert.Empty(t, err)
			}
		})
	}
}
