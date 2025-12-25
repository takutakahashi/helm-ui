package helm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/helm-version-manager/api/internal/model"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Client struct {
	settings   *cli.EnvSettings
	repoFile   string
	mu         sync.RWMutex
}

func NewClient() (*Client, error) {
	settings := cli.New()

	if _, err := config.GetConfig(); err != nil {
		return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	return &Client{
		settings: settings,
		repoFile: settings.RepositoryConfig,
	}, nil
}

func (c *Client) getActionConfig(namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		home, _ := os.UserHomeDir()
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	if err := actionConfig.Init(
		c.settings.RESTClientGetter(),
		namespace,
		os.Getenv("HELM_DRIVER"),
		func(format string, v ...interface{}) {},
	); err != nil {
		return nil, fmt.Errorf("failed to init action config: %w", err)
	}

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

	chartPath, err := c.locateChart(chartName, req.ChartVersion)
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

	return c.searchChartVersions(chartName)
}

func (c *Client) searchChartVersions(chartName string) ([]model.ChartVersion, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	repoFile, err := repo.LoadFile(c.repoFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load repo file: %w", err)
	}

	var versions []model.ChartVersion

	for _, repoEntry := range repoFile.Repositories {
		indexFile, err := repo.LoadIndexFile(c.settings.RepositoryCache + "/" + repoEntry.Name + "-index.yaml")
		if err != nil {
			continue
		}

		for name, chartVersions := range indexFile.Entries {
			if name == chartName || strings.HasSuffix(chartName, "/"+name) {
				for _, cv := range chartVersions {
					versions = append(versions, model.ChartVersion{
						Version:     cv.Version,
						AppVersion:  cv.AppVersion,
						Description: cv.Description,
					})
				}
			}
		}
	}

	return versions, nil
}

func (c *Client) locateChart(chartName, version string) (string, error) {
	client := action.NewPull()
	client.Settings = c.settings
	client.Version = version

	if version != "" {
		client.Version = version
	}

	chartPath, err := client.LocateChart(chartName, c.settings)
	if err != nil {
		return "", fmt.Errorf("failed to locate chart %s version %s: %w", chartName, version, err)
	}

	return chartPath, nil
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
