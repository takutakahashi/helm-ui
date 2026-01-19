package helm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/helm-version-manager/api/internal/model"
	"github.com/helm-version-manager/api/internal/storage"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"oras.land/oras-go/v2/registry/remote"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Client struct {
	settings      *cli.EnvSettings
	registryStore *storage.RegistryStore
	mu            sync.RWMutex
}

func NewClient(store *storage.RegistryStore) (*Client, error) {
	settings := cli.New()

	if _, err := config.GetConfig(); err != nil {
		return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	return &Client{
		settings:      settings,
		registryStore: store,
	}, nil
}

// buildConfigFlags creates ConfigFlags with the specified namespace.
// This ensures the namespace is explicitly set, ignoring HELM_NAMESPACE env var.
func (c *Client) buildConfigFlags(namespace string) *genericclioptions.ConfigFlags {
	configFlags := genericclioptions.NewConfigFlags(true)
	configFlags.Namespace = &namespace

	// Only set KubeConfig if KUBECONFIG env var is explicitly set
	// Otherwise, let genericclioptions use its default behavior (in-cluster config or ~/.kube/config)
	if kubeconfigPath := os.Getenv("KUBECONFIG"); kubeconfigPath != "" {
		configFlags.KubeConfig = &kubeconfigPath
	}

	return configFlags
}

func (c *Client) getActionConfig(namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	// Use explicit namespace via buildConfigFlags
	// instead of c.settings.RESTClientGetter() which may use helm-ui's namespace
	configFlags := c.buildConfigFlags(namespace)

	if err := actionConfig.Init(
		configFlags,
		namespace,
		os.Getenv("HELM_DRIVER"),
		func(format string, v ...interface{}) {},
	); err != nil {
		return nil, fmt.Errorf("failed to init action config: %w", err)
	}

	registryClient, err := registry.NewClient(
		registry.ClientOptCredentialsFile(c.settings.RegistryConfig),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry client: %w", err)
	}
	actionConfig.RegistryClient = registryClient

	return actionConfig, nil
}

func (c *Client) ListReleases() ([]model.Release, error) {
	actionConfig, err := c.getActionConfig("")
	if err != nil {
		return nil, err
	}

	listAction := action.NewList(actionConfig)
	listAction.AllNamespaces = true
	listAction.All = true

	releases, err := listAction.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %w", err)
	}

	result := make([]model.Release, 0, len(releases))
	for _, r := range releases {
		result = append(result, toModelRelease(r))
	}

	return result, nil
}

func (c *Client) GetRelease(namespace, name string) (*model.Release, error) {
	actionConfig, err := c.getActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	getAction := action.NewGet(actionConfig)
	r, err := getAction.Run(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get release %s/%s: %w", namespace, name, err)
	}

	result := toModelRelease(r)
	return &result, nil
}

func (c *Client) GetReleaseHistory(namespace, name string) ([]model.ReleaseHistory, error) {
	actionConfig, err := c.getActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	historyAction := action.NewHistory(actionConfig)
	historyAction.Max = 10

	releases, err := historyAction.Run(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get history for %s/%s: %w", namespace, name, err)
	}

	result := make([]model.ReleaseHistory, 0, len(releases))
	for _, r := range releases {
		result = append(result, model.ReleaseHistory{
			Revision:    r.Version,
			Updated:     r.Info.LastDeployed.Time,
			Status:      string(r.Info.Status),
			Chart:       r.Chart.Metadata.Version,
			AppVersion:  r.Chart.Metadata.AppVersion,
			Description: r.Info.Description,
		})
	}

	return result, nil
}

func (c *Client) UpgradeRelease(namespace, name string, req model.VersionUpgradeRequest) (*model.Release, error) {
	actionConfig, err := c.getActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	getAction := action.NewGet(actionConfig)
	currentRelease, err := getAction.Run(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get current release: %w", err)
	}

	chartName := currentRelease.Chart.Metadata.Name

	mapping, err := c.registryStore.GetMapping(context.Background(), namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get registry mapping: %w", err)
	}
	if mapping == nil {
		return nil, fmt.Errorf("registry mapping not found for release %s/%s, please set registry first", namespace, name)
	}

	chartPath, err := c.locateChart(actionConfig, mapping.Registry, chartName, req.ChartVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart: %w", err)
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	upgradeAction := action.NewUpgrade(actionConfig)
	upgradeAction.Namespace = namespace
	upgradeAction.ReuseValues = true

	vals := currentRelease.Config
	if req.Values != nil {
		for k, v := range req.Values {
			vals[k] = v
		}
	}

	r, err := upgradeAction.Run(name, chart, vals)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade release: %w", err)
	}

	result := toModelRelease(r)
	return &result, nil
}

func (c *Client) GetAvailableVersions(namespace, name string) ([]model.ChartVersion, error) {
	actionConfig, err := c.getActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	getAction := action.NewGet(actionConfig)
	currentRelease, err := getAction.Run(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get current release: %w", err)
	}

	chartName := currentRelease.Chart.Metadata.Name

	mapping, err := c.registryStore.GetMapping(context.Background(), namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get registry mapping: %w", err)
	}
	if mapping == nil {
		return nil, fmt.Errorf("registry mapping not found for release %s/%s, please set registry first", namespace, name)
	}

	return c.searchChartVersions(mapping.Registry, chartName)
}

func (c *Client) searchChartVersions(reg, chartName string) ([]model.ChartVersion, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	reg = strings.TrimPrefix(reg, "oci://")
	repoRef := fmt.Sprintf("%s/%s", reg, chartName)

	repo, err := remote.NewRepository(repoRef)
	if err != nil {
		return nil, fmt.Errorf("failed to create OCI repository client: %w", err)
	}

	ctx := context.Background()
	var tags []string

	err = repo.Tags(ctx, "", func(t []string) error {
		tags = append(tags, t...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list tags from OCI registry: %w", err)
	}

	// Reverse the order (newest first)
	for i, j := 0, len(tags)-1; i < j; i, j = i+1, j-1 {
		tags[i], tags[j] = tags[j], tags[i]
	}

	// Limit to 10 latest versions
	if len(tags) > 10 {
		tags = tags[:10]
	}

	versions := make([]model.ChartVersion, 0, len(tags))
	for _, tag := range tags {
		versions = append(versions, model.ChartVersion{
			Version: tag,
		})
	}

	return versions, nil
}

func (c *Client) locateChart(actionConfig *action.Configuration, reg, chartName, version string) (string, error) {
	reg = strings.TrimPrefix(reg, "oci://")
	ref := fmt.Sprintf("%s/%s:%s", reg, chartName, version)

	registryClient, err := registry.NewClient(
		registry.ClientOptCredentialsFile(c.settings.RegistryConfig),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create registry client: %w", err)
	}

	result, err := registryClient.Pull(ref)
	if err != nil {
		return "", fmt.Errorf("failed to pull chart %s: %w", ref, err)
	}

	// Save chart to cache directory
	cacheDir := c.settings.RepositoryCache
	if cacheDir == "" {
		cacheDir = os.TempDir()
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory %s: %w", cacheDir, err)
	}
	chartFileName := fmt.Sprintf("%s-%s.tgz", chartName, version)
	chartPath := filepath.Join(cacheDir, chartFileName)

	if err := os.WriteFile(chartPath, result.Chart.Data, 0644); err != nil {
		return "", fmt.Errorf("failed to write chart to %s: %w", chartPath, err)
	}

	return chartPath, nil
}

func (c *Client) GetReleaseValues(namespace, name string) (map[string]any, error) {
	actionConfig, err := c.getActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	getAction := action.NewGet(actionConfig)
	r, err := getAction.Run(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get release %s/%s: %w", namespace, name, err)
	}

	return r.Config, nil
}

func (c *Client) UpdateReleaseValues(namespace, name string, values map[string]any) (*model.Release, error) {
	actionConfig, err := c.getActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	getAction := action.NewGet(actionConfig)
	currentRelease, err := getAction.Run(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get current release: %w", err)
	}

	chartName := currentRelease.Chart.Metadata.Name
	chartVersion := currentRelease.Chart.Metadata.Version

	mapping, err := c.registryStore.GetMapping(context.Background(), namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get registry mapping: %w", err)
	}
	if mapping == nil {
		return nil, fmt.Errorf("registry mapping not found for release %s/%s, please set registry first", namespace, name)
	}

	chartPath, err := c.locateChart(actionConfig, mapping.Registry, chartName, chartVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart: %w", err)
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	upgradeAction := action.NewUpgrade(actionConfig)
	upgradeAction.Namespace = namespace
	upgradeAction.ReuseValues = true

	// Merge new values with existing values
	vals := currentRelease.Config
	for k, v := range values {
		vals[k] = v
	}

	r, err := upgradeAction.Run(name, chart, vals)
	if err != nil {
		return nil, fmt.Errorf("failed to update release values: %w", err)
	}

	result := toModelRelease(r)
	return &result, nil
}

func (c *Client) RollbackRelease(namespace, name string, revision int) (*model.Release, error) {
	actionConfig, err := c.getActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	rollbackAction := action.NewRollback(actionConfig)
	rollbackAction.Version = revision

	if err := rollbackAction.Run(name); err != nil {
		return nil, fmt.Errorf("failed to rollback release %s/%s to revision %d: %w", namespace, name, revision, err)
	}

	// Get the updated release after rollback
	getAction := action.NewGet(actionConfig)
	r, err := getAction.Run(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get release after rollback: %w", err)
	}

	result := toModelRelease(r)
	return &result, nil
}

func toModelRelease(r *release.Release) model.Release {
	return model.Release{
		Name:         r.Name,
		Namespace:    r.Namespace,
		Chart:        r.Chart.Metadata.Name,
		ChartVersion: r.Chart.Metadata.Version,
		AppVersion:   r.Chart.Metadata.AppVersion,
		Status:       string(r.Info.Status),
		Updated:      r.Info.LastDeployed.Time,
		Revision:     r.Version,
	}
}
