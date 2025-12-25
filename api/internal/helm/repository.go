package helm

import (
	"fmt"
	"os"

	"github.com/helm-version-manager/api/internal/model"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

func (c *Client) ListRepositories() ([]model.Repository, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	repoFile, err := repo.LoadFile(c.repoFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.Repository{}, nil
		}
		return nil, fmt.Errorf("failed to load repo file: %w", err)
	}

	result := make([]model.Repository, 0, len(repoFile.Repositories))
	for _, r := range repoFile.Repositories {
		result = append(result, model.Repository{
			Name: r.Name,
			URL:  r.URL,
		})
	}

	return result, nil
}

func (c *Client) AddRepository(req model.AddRepositoryRequest) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	repoFile, err := repo.LoadFile(c.repoFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load repo file: %w", err)
		}
		repoFile = repo.NewFile()
	}

	if repoFile.Has(req.Name) {
		return fmt.Errorf("repository %s already exists", req.Name)
	}

	entry := &repo.Entry{
		Name: req.Name,
		URL:  req.URL,
	}

	chartRepo, err := repo.NewChartRepository(entry, getter.All(c.settings))
	if err != nil {
		return fmt.Errorf("failed to create chart repository: %w", err)
	}

	if _, err := chartRepo.DownloadIndexFile(); err != nil {
		return fmt.Errorf("failed to download index file: %w", err)
	}

	repoFile.Update(entry)

	if err := repoFile.WriteFile(c.repoFile, 0644); err != nil {
		return fmt.Errorf("failed to write repo file: %w", err)
	}

	return nil
}

func (c *Client) RemoveRepository(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	repoFile, err := repo.LoadFile(c.repoFile)
	if err != nil {
		return fmt.Errorf("failed to load repo file: %w", err)
	}

	if !repoFile.Has(name) {
		return fmt.Errorf("repository %s not found", name)
	}

	if !repoFile.Remove(name) {
		return fmt.Errorf("failed to remove repository %s", name)
	}

	if err := repoFile.WriteFile(c.repoFile, 0644); err != nil {
		return fmt.Errorf("failed to write repo file: %w", err)
	}

	indexFile := c.settings.RepositoryCache + "/" + name + "-index.yaml"
	if err := os.Remove(indexFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove index file: %w", err)
	}

	return nil
}

func (c *Client) UpdateRepository(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	repoFile, err := repo.LoadFile(c.repoFile)
	if err != nil {
		return fmt.Errorf("failed to load repo file: %w", err)
	}

	var entry *repo.Entry
	for _, r := range repoFile.Repositories {
		if r.Name == name {
			entry = r
			break
		}
	}

	if entry == nil {
		return fmt.Errorf("repository %s not found", name)
	}

	chartRepo, err := repo.NewChartRepository(entry, getter.All(c.settings))
	if err != nil {
		return fmt.Errorf("failed to create chart repository: %w", err)
	}

	if _, err := chartRepo.DownloadIndexFile(); err != nil {
		return fmt.Errorf("failed to update repository index: %w", err)
	}

	return nil
}
