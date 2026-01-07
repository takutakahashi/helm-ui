package mcp

import (
	"context"

	"github.com/helm-version-manager/api/internal/model"
)

// HelmClient defines the interface for Helm operations
type HelmClient interface {
	ListReleases() ([]model.Release, error)
	GetRelease(namespace, name string) (*model.Release, error)
	GetAvailableVersions(namespace, name string) ([]model.ChartVersion, error)
	UpgradeRelease(namespace, name string, req model.VersionUpgradeRequest) (*model.Release, error)
	GetReleaseHistory(namespace, name string) ([]model.ReleaseHistory, error)
	GetReleaseValues(namespace, name string) (map[string]any, error)
	UpdateReleaseValues(namespace, name string, values map[string]any) (*model.Release, error)
}

// RegistryStore defines the interface for registry mapping storage
type RegistryStore interface {
	GetMapping(ctx context.Context, namespace, releaseName string) (*model.RegistryMapping, error)
	SetMapping(ctx context.Context, mapping model.RegistryMapping) error
	DeleteMapping(ctx context.Context, namespace, releaseName string) error
	ListMappings(ctx context.Context) ([]model.RegistryMapping, error)
}
