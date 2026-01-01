package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/helm-version-manager/api/internal/model"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Mock implementations

type mockHelmClient struct {
	releases       []model.Release
	releaseDetails map[string]*model.Release
	versions       map[string][]model.ChartVersion
	history        map[string][]model.ReleaseHistory
	listErr        error
	getErr         error
	versionsErr    error
	upgradeErr     error
	historyErr     error
}

func (m *mockHelmClient) ListReleases() ([]model.Release, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.releases, nil
}

func (m *mockHelmClient) GetRelease(namespace, name string) (*model.Release, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	key := namespace + "/" + name
	if r, ok := m.releaseDetails[key]; ok {
		return r, nil
	}
	return nil, nil
}

func (m *mockHelmClient) GetAvailableVersions(namespace, name string) ([]model.ChartVersion, error) {
	if m.versionsErr != nil {
		return nil, m.versionsErr
	}
	key := namespace + "/" + name
	return m.versions[key], nil
}

func (m *mockHelmClient) UpgradeRelease(namespace, name string, req model.VersionUpgradeRequest) (*model.Release, error) {
	if m.upgradeErr != nil {
		return nil, m.upgradeErr
	}
	key := namespace + "/" + name
	if r, ok := m.releaseDetails[key]; ok {
		upgraded := *r
		upgraded.ChartVersion = req.ChartVersion
		upgraded.Revision = r.Revision + 1
		return &upgraded, nil
	}
	return nil, nil
}

func (m *mockHelmClient) GetReleaseHistory(namespace, name string) ([]model.ReleaseHistory, error) {
	if m.historyErr != nil {
		return nil, m.historyErr
	}
	key := namespace + "/" + name
	return m.history[key], nil
}

type mockRegistryStore struct {
	mappings  map[string]*model.RegistryMapping
	getErr    error
	setErr    error
	deleteErr error
}

func (m *mockRegistryStore) GetMapping(ctx context.Context, namespace, releaseName string) (*model.RegistryMapping, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	key := namespace + "/" + releaseName
	return m.mappings[key], nil
}

func (m *mockRegistryStore) SetMapping(ctx context.Context, mapping model.RegistryMapping) error {
	if m.setErr != nil {
		return m.setErr
	}
	key := mapping.Namespace + "/" + mapping.ReleaseName
	m.mappings[key] = &mapping
	return nil
}

func (m *mockRegistryStore) DeleteMapping(ctx context.Context, namespace, releaseName string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	key := namespace + "/" + releaseName
	delete(m.mappings, key)
	return nil
}

func (m *mockRegistryStore) ListMappings(ctx context.Context) ([]model.RegistryMapping, error) {
	result := make([]model.RegistryMapping, 0, len(m.mappings))
	for _, mapping := range m.mappings {
		result = append(result, *mapping)
	}
	return result, nil
}

// Helper function to create test server
func newTestServer(helmClient *mockHelmClient, registryStore *mockRegistryStore) *Server {
	return NewServer(helmClient, registryStore)
}

// Tests

func TestNewServer(t *testing.T) {
	helmClient := &mockHelmClient{}
	registryStore := &mockRegistryStore{mappings: make(map[string]*model.RegistryMapping)}

	server := NewServer(helmClient, registryStore)

	if server == nil {
		t.Fatal("expected server to be non-nil")
	}
	if server.mcpServer == nil {
		t.Fatal("expected mcpServer to be non-nil")
	}
	if server.helmClient == nil {
		t.Fatal("expected helmClient to be non-nil")
	}
	if server.registryStore == nil {
		t.Fatal("expected registryStore to be non-nil")
	}
}

func TestHandleListReleases(t *testing.T) {
	now := time.Now()
	releases := []model.Release{
		{Name: "release1", Namespace: "default", Chart: "mychart", ChartVersion: "1.0.0", Status: "deployed", Updated: now},
		{Name: "release2", Namespace: "kube-system", Chart: "anotherchart", ChartVersion: "2.0.0", Status: "deployed", Updated: now},
	}

	helmClient := &mockHelmClient{
		releases: releases,
	}
	registryStore := &mockRegistryStore{
		mappings: map[string]*model.RegistryMapping{
			"default/release1": {Namespace: "default", ReleaseName: "release1", Registry: "oci://example.com/charts"},
		},
	}

	server := newTestServer(helmClient, registryStore)

	ctx := context.Background()

	t.Run("list all releases", func(t *testing.T) {
		result, output, err := server.handleListReleases(ctx, &mcp.CallToolRequest{}, ListReleasesInput{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Fatal("expected result to be nil for success")
		}
		if len(output.Releases) != 2 {
			t.Errorf("expected 2 releases, got %d", len(output.Releases))
		}
	})

	t.Run("filter by namespace", func(t *testing.T) {
		result, output, err := server.handleListReleases(ctx, &mcp.CallToolRequest{}, ListReleasesInput{Namespace: "default"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Fatal("expected result to be nil for success")
		}
		if len(output.Releases) != 1 {
			t.Errorf("expected 1 release, got %d", len(output.Releases))
		}
		if output.Releases[0].Name != "release1" {
			t.Errorf("expected release1, got %s", output.Releases[0].Name)
		}
	})

	t.Run("filter by hasRegistry true", func(t *testing.T) {
		hasRegistry := true
		result, output, err := server.handleListReleases(ctx, &mcp.CallToolRequest{}, ListReleasesInput{HasRegistry: &hasRegistry})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Fatal("expected result to be nil for success")
		}
		if len(output.Releases) != 1 {
			t.Errorf("expected 1 release with registry, got %d", len(output.Releases))
		}
		if output.Releases[0].Name != "release1" {
			t.Errorf("expected release1, got %s", output.Releases[0].Name)
		}
	})

	t.Run("filter by hasRegistry false", func(t *testing.T) {
		hasRegistry := false
		result, output, err := server.handleListReleases(ctx, &mcp.CallToolRequest{}, ListReleasesInput{HasRegistry: &hasRegistry})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Fatal("expected result to be nil for success")
		}
		if len(output.Releases) != 1 {
			t.Errorf("expected 1 release without registry, got %d", len(output.Releases))
		}
		if output.Releases[0].Name != "release2" {
			t.Errorf("expected release2, got %s", output.Releases[0].Name)
		}
	})
}

func TestHandleGetRelease(t *testing.T) {
	now := time.Now()
	release := &model.Release{
		Name:         "myrelease",
		Namespace:    "default",
		Chart:        "mychart",
		ChartVersion: "1.0.0",
		Status:       "deployed",
		Updated:      now,
		Revision:     3,
	}

	helmClient := &mockHelmClient{
		releaseDetails: map[string]*model.Release{
			"default/myrelease": release,
		},
	}
	registryStore := &mockRegistryStore{
		mappings: make(map[string]*model.RegistryMapping),
	}

	server := newTestServer(helmClient, registryStore)
	ctx := context.Background()

	t.Run("get existing release", func(t *testing.T) {
		result, output, err := server.handleGetRelease(ctx, &mcp.CallToolRequest{}, ReleaseInput{
			Namespace: "default",
			Name:      "myrelease",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Fatal("expected result to be nil for success")
		}
		if output.Release == nil {
			t.Fatal("expected release to be non-nil")
		}
		if output.Release.Name != "myrelease" {
			t.Errorf("expected myrelease, got %s", output.Release.Name)
		}
	})

	t.Run("missing namespace", func(t *testing.T) {
		_, _, err := server.handleGetRelease(ctx, &mcp.CallToolRequest{}, ReleaseInput{
			Name: "myrelease",
		})
		if err == nil {
			t.Fatal("expected error for missing namespace")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		_, _, err := server.handleGetRelease(ctx, &mcp.CallToolRequest{}, ReleaseInput{
			Namespace: "default",
		})
		if err == nil {
			t.Fatal("expected error for missing name")
		}
	})
}

func TestHandleGetAvailableVersions(t *testing.T) {
	versions := []model.ChartVersion{
		{Version: "1.2.0"},
		{Version: "1.1.0"},
		{Version: "1.0.0"},
	}

	helmClient := &mockHelmClient{
		versions: map[string][]model.ChartVersion{
			"default/myrelease": versions,
		},
	}
	registryStore := &mockRegistryStore{mappings: make(map[string]*model.RegistryMapping)}

	server := newTestServer(helmClient, registryStore)
	ctx := context.Background()

	t.Run("get versions", func(t *testing.T) {
		result, output, err := server.handleGetAvailableVersions(ctx, &mcp.CallToolRequest{}, ReleaseInput{
			Namespace: "default",
			Name:      "myrelease",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Fatal("expected result to be nil for success")
		}
		if len(output.Versions) != 3 {
			t.Errorf("expected 3 versions, got %d", len(output.Versions))
		}
	})
}

func TestHandleUpgradeRelease(t *testing.T) {
	now := time.Now()
	release := &model.Release{
		Name:         "myrelease",
		Namespace:    "default",
		Chart:        "mychart",
		ChartVersion: "1.0.0",
		Status:       "deployed",
		Updated:      now,
		Revision:     1,
	}

	helmClient := &mockHelmClient{
		releaseDetails: map[string]*model.Release{
			"default/myrelease": release,
		},
	}
	registryStore := &mockRegistryStore{mappings: make(map[string]*model.RegistryMapping)}

	server := newTestServer(helmClient, registryStore)
	ctx := context.Background()

	t.Run("upgrade release", func(t *testing.T) {
		result, output, err := server.handleUpgradeRelease(ctx, &mcp.CallToolRequest{}, UpgradeInput{
			Namespace:    "default",
			Name:         "myrelease",
			ChartVersion: "1.1.0",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Fatal("expected result to be nil for success")
		}
		if output.Release == nil {
			t.Fatal("expected release to be non-nil")
		}
		if output.Release.ChartVersion != "1.1.0" {
			t.Errorf("expected version 1.1.0, got %s", output.Release.ChartVersion)
		}
		if output.Release.Revision != 2 {
			t.Errorf("expected revision 2, got %d", output.Release.Revision)
		}
	})

	t.Run("missing chart_version", func(t *testing.T) {
		_, _, err := server.handleUpgradeRelease(ctx, &mcp.CallToolRequest{}, UpgradeInput{
			Namespace: "default",
			Name:      "myrelease",
		})
		if err == nil {
			t.Fatal("expected error for missing chart_version")
		}
	})
}

func TestHandleGetReleaseHistory(t *testing.T) {
	now := time.Now()
	history := []model.ReleaseHistory{
		{Revision: 3, Updated: now, Status: "deployed", Chart: "1.2.0"},
		{Revision: 2, Updated: now.Add(-time.Hour), Status: "superseded", Chart: "1.1.0"},
		{Revision: 1, Updated: now.Add(-2 * time.Hour), Status: "superseded", Chart: "1.0.0"},
	}

	helmClient := &mockHelmClient{
		history: map[string][]model.ReleaseHistory{
			"default/myrelease": history,
		},
	}
	registryStore := &mockRegistryStore{mappings: make(map[string]*model.RegistryMapping)}

	server := newTestServer(helmClient, registryStore)
	ctx := context.Background()

	t.Run("get history", func(t *testing.T) {
		result, output, err := server.handleGetReleaseHistory(ctx, &mcp.CallToolRequest{}, ReleaseInput{
			Namespace: "default",
			Name:      "myrelease",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Fatal("expected result to be nil for success")
		}
		if len(output.History) != 3 {
			t.Errorf("expected 3 history entries, got %d", len(output.History))
		}
	})
}

func TestHandleGetRegistry(t *testing.T) {
	mapping := &model.RegistryMapping{
		Namespace:   "default",
		ReleaseName: "myrelease",
		ChartName:   "mychart",
		Registry:    "oci://example.com/charts",
	}

	helmClient := &mockHelmClient{}
	registryStore := &mockRegistryStore{
		mappings: map[string]*model.RegistryMapping{
			"default/myrelease": mapping,
		},
	}

	server := newTestServer(helmClient, registryStore)
	ctx := context.Background()

	t.Run("get existing mapping", func(t *testing.T) {
		result, output, err := server.handleGetRegistry(ctx, &mcp.CallToolRequest{}, RegistryInput{
			Namespace: "default",
			Name:      "myrelease",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Fatal("expected result to be nil for success")
		}
		if output.Mapping == nil {
			t.Fatal("expected mapping to be non-nil")
		}
		if output.Mapping.Registry != "oci://example.com/charts" {
			t.Errorf("expected oci://example.com/charts, got %s", output.Mapping.Registry)
		}
	})

	t.Run("get non-existing mapping", func(t *testing.T) {
		result, output, err := server.handleGetRegistry(ctx, &mcp.CallToolRequest{}, RegistryInput{
			Namespace: "default",
			Name:      "nonexistent",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Fatal("expected result to be nil for success")
		}
		if output.Mapping != nil {
			t.Fatal("expected mapping to be nil")
		}
		if output.Message == "" {
			t.Fatal("expected message for non-existing mapping")
		}
	})
}

func TestHandleSetRegistry(t *testing.T) {
	release := &model.Release{
		Name:         "myrelease",
		Namespace:    "default",
		Chart:        "mychart",
		ChartVersion: "1.0.0",
	}

	helmClient := &mockHelmClient{
		releaseDetails: map[string]*model.Release{
			"default/myrelease": release,
		},
	}
	registryStore := &mockRegistryStore{
		mappings: make(map[string]*model.RegistryMapping),
	}

	server := newTestServer(helmClient, registryStore)
	ctx := context.Background()

	t.Run("set registry", func(t *testing.T) {
		result, output, err := server.handleSetRegistry(ctx, &mcp.CallToolRequest{}, SetRegistryInput{
			Namespace: "default",
			Name:      "myrelease",
			Registry:  "oci://example.com/charts",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Fatal("expected result to be nil for success")
		}
		if output.Mapping == nil {
			t.Fatal("expected mapping to be non-nil")
		}
		if output.Mapping.Registry != "oci://example.com/charts" {
			t.Errorf("expected oci://example.com/charts, got %s", output.Mapping.Registry)
		}
		if output.Mapping.ChartName != "mychart" {
			t.Errorf("expected mychart, got %s", output.Mapping.ChartName)
		}

		// Verify mapping was stored
		stored := registryStore.mappings["default/myrelease"]
		if stored == nil {
			t.Fatal("expected mapping to be stored")
		}
	})

	t.Run("missing registry", func(t *testing.T) {
		_, _, err := server.handleSetRegistry(ctx, &mcp.CallToolRequest{}, SetRegistryInput{
			Namespace: "default",
			Name:      "myrelease",
		})
		if err == nil {
			t.Fatal("expected error for missing registry")
		}
	})
}

func TestHandleDeleteRegistry(t *testing.T) {
	helmClient := &mockHelmClient{}
	registryStore := &mockRegistryStore{
		mappings: map[string]*model.RegistryMapping{
			"default/myrelease": {
				Namespace:   "default",
				ReleaseName: "myrelease",
				Registry:    "oci://example.com/charts",
			},
		},
	}

	server := newTestServer(helmClient, registryStore)
	ctx := context.Background()

	t.Run("delete registry", func(t *testing.T) {
		result, output, err := server.handleDeleteRegistry(ctx, &mcp.CallToolRequest{}, RegistryInput{
			Namespace: "default",
			Name:      "myrelease",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Fatal("expected result to be nil for success")
		}
		if !output.Success {
			t.Fatal("expected success to be true")
		}

		// Verify mapping was deleted
		if _, exists := registryStore.mappings["default/myrelease"]; exists {
			t.Fatal("expected mapping to be deleted")
		}
	})
}

func TestMCPServer(t *testing.T) {
	helmClient := &mockHelmClient{}
	registryStore := &mockRegistryStore{mappings: make(map[string]*model.RegistryMapping)}

	server := NewServer(helmClient, registryStore)

	mcpServer := server.MCPServer()
	if mcpServer == nil {
		t.Fatal("expected MCPServer to be non-nil")
	}
}
