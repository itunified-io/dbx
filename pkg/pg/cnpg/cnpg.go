// Package cnpg provides CloudNativePG cluster management tools.
package cnpg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ClusterStatusResult represents CNPG cluster status.
type ClusterStatusResult struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	Phase          string `json:"phase"`
	Instances      int64  `json:"instances"`
	ReadyInstances int64  `json:"ready_instances"`
	PrimaryPod     string `json:"primary_pod"`
}

// ClusterSummary represents a brief cluster listing entry.
type ClusterSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Instances int64  `json:"instances"`
	Phase     string `json:"phase"`
}

// BackupInfo represents a CNPG backup.
type BackupInfo struct {
	Name       string     `json:"name"`
	Cluster    string     `json:"cluster"`
	Phase      string     `json:"phase"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	StoppedAt  *time.Time `json:"stopped_at,omitempty"`
}

// ClusterStatus retrieves status for a specific CNPG cluster.
func ClusterStatus(ctx context.Context, k8s *pginternal.K8sClient, clusterName string) (*ClusterStatusResult, error) {
	cluster, err := k8s.GetCluster(ctx, clusterName)
	if err != nil {
		return nil, fmt.Errorf("cnpg cluster status: %w", err)
	}

	result := &ClusterStatusResult{
		Name:      cluster.GetName(),
		Namespace: cluster.GetNamespace(),
	}

	if phase, ok, _ := unstructured.NestedString(cluster.Object, "status", "phase"); ok {
		result.Phase = phase
	}
	if instances, ok, _ := unstructured.NestedInt64(cluster.Object, "status", "instances"); ok {
		result.Instances = instances
	}
	if ready, ok, _ := unstructured.NestedInt64(cluster.Object, "status", "readyInstances"); ok {
		result.ReadyInstances = ready
	}
	if primary, ok, _ := unstructured.NestedString(cluster.Object, "status", "currentPrimary"); ok {
		result.PrimaryPod = primary
	}

	return result, nil
}

// ClusterList lists all CNPG clusters in the namespace.
func ClusterList(ctx context.Context, k8s *pginternal.K8sClient) ([]ClusterSummary, error) {
	list, err := k8s.ListClusters(ctx)
	if err != nil {
		return nil, fmt.Errorf("cnpg cluster list: %w", err)
	}

	var results []ClusterSummary
	for _, item := range list.Items {
		summary := ClusterSummary{
			Name:      item.GetName(),
			Namespace: item.GetNamespace(),
		}
		if instances, ok, _ := unstructured.NestedInt64(item.Object, "spec", "instances"); ok {
			summary.Instances = instances
		}
		if phase, ok, _ := unstructured.NestedString(item.Object, "status", "phase"); ok {
			summary.Phase = phase
		}
		results = append(results, summary)
	}
	return results, nil
}

// BackupList lists backups for a cluster.
func BackupList(ctx context.Context, k8s *pginternal.K8sClient, clusterName string) ([]BackupInfo, error) {
	list, err := k8s.ListBackups(ctx, "cnpg.io/cluster="+clusterName)
	if err != nil {
		return nil, fmt.Errorf("cnpg backup list: %w", err)
	}

	var results []BackupInfo
	for _, item := range list.Items {
		info := BackupInfo{
			Name:    item.GetName(),
			Cluster: clusterName,
		}
		if phase, ok, _ := unstructured.NestedString(item.Object, "status", "phase"); ok {
			info.Phase = phase
		}
		results = append(results, info)
	}
	return results, nil
}

// BackupTrigger creates a new backup for a cluster. Confirm-gated.
func BackupTrigger(ctx context.Context, k8s *pginternal.K8sClient, clusterName string, confirm bool) (*BackupInfo, error) {
	if !confirm {
		return nil, fmt.Errorf("confirm gate: set confirm=true to trigger backup")
	}

	backup := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "postgresql.cnpg.io/v1",
			"kind":       "Backup",
			"metadata": map[string]any{
				"name":      fmt.Sprintf("%s-backup-%d", clusterName, time.Now().Unix()),
				"namespace": k8s.Namespace(),
			},
			"spec": map[string]any{
				"cluster": map[string]any{
					"name": clusterName,
				},
			},
		},
	}

	result, err := k8s.CreateBackup(ctx, backup)
	if err != nil {
		return nil, fmt.Errorf("cnpg backup trigger: %w", err)
	}
	return &BackupInfo{Name: result.GetName(), Cluster: clusterName, Phase: "pending"}, nil
}

// ClusterScale changes the number of instances. Confirm-gated.
func ClusterScale(ctx context.Context, k8s *pginternal.K8sClient, clusterName string, instances int, confirm bool) error {
	if !confirm {
		return fmt.Errorf("confirm gate: set confirm=true to scale cluster")
	}

	patch := map[string]any{
		"spec": map[string]any{
			"instances": instances,
		},
	}
	patchBytes, _ := json.Marshal(patch)
	if err := k8s.PatchCluster(ctx, clusterName, patchBytes); err != nil {
		return fmt.Errorf("cnpg cluster scale: %w", err)
	}
	return nil
}

// ClusterRestart triggers a rolling restart. Confirm-gated.
func ClusterRestart(ctx context.Context, k8s *pginternal.K8sClient, clusterName string, confirm bool) error {
	if !confirm {
		return fmt.Errorf("confirm gate: set confirm=true to restart cluster")
	}

	patch := map[string]any{
		"metadata": map[string]any{
			"annotations": map[string]any{
				"cnpg.io/restartedAt": time.Now().Format(time.RFC3339),
			},
		},
	}
	patchBytes, _ := json.Marshal(patch)
	if err := k8s.PatchCluster(ctx, clusterName, patchBytes); err != nil {
		return fmt.Errorf("cnpg cluster restart: %w", err)
	}
	return nil
}
