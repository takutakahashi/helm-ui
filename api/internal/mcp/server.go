package mcp

import (
	"context"
	"fmt"

	"github.com/helm-version-manager/api/internal/model"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server with Helm functionality
type Server struct {
	mcpServer     *mcp.Server
	helmClient    HelmClient
	registryStore RegistryStore
}

// Input/Output types for MCP tools

type ListReleasesInput struct {
	Namespace   string `json:"namespace,omitempty" jsonschema:"Filter by namespace (optional)"`
	HasRegistry *bool  `json:"has_registry,omitempty" jsonschema:"Filter by whether the release has a registry mapping configured (optional)"`
}

type ListReleasesOutput struct {
	Releases []model.Release `json:"releases"`
}

type ReleaseInput struct {
	Namespace string `json:"namespace" jsonschema:"The namespace of the release"`
	Name      string `json:"name" jsonschema:"The name of the release"`
}

type ReleaseOutput struct {
	Release *model.Release `json:"release"`
}

type VersionsOutput struct {
	Versions []model.ChartVersion `json:"versions"`
}

type UpgradeInput struct {
	Namespace    string `json:"namespace" jsonschema:"The namespace of the release"`
	Name         string `json:"name" jsonschema:"The name of the release"`
	ChartVersion string `json:"chart_version" jsonschema:"The target chart version to upgrade to"`
}

type HistoryOutput struct {
	History []model.ReleaseHistory `json:"history"`
}

type RegistryInput struct {
	Namespace string `json:"namespace" jsonschema:"The namespace of the release"`
	Name      string `json:"name" jsonschema:"The name of the release"`
}

type RegistryOutput struct {
	Mapping *model.RegistryMapping `json:"mapping,omitempty"`
	Message string                 `json:"message,omitempty"`
}

type SetRegistryInput struct {
	Namespace string `json:"namespace" jsonschema:"The namespace of the release"`
	Name      string `json:"name" jsonschema:"The name of the release"`
	Registry  string `json:"registry" jsonschema:"The OCI registry URL (e.g. oci://ghcr.io/myorg/charts)"`
}

type DeleteRegistryOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ValuesOutput struct {
	Values map[string]any `json:"values"`
}

type UpdateValuesInput struct {
	Namespace string         `json:"namespace" jsonschema:"The namespace of the release"`
	Name      string         `json:"name" jsonschema:"The name of the release"`
	Values    map[string]any `json:"values" jsonschema:"The new values to set for the release"`
}

// NewServer creates a new MCP server with Helm tools
func NewServer(helmClient HelmClient, registryStore RegistryStore) *Server {
	s := &Server{
		helmClient:    helmClient,
		registryStore: registryStore,
	}

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "helm-version-manager",
			Version: "1.0.0",
		},
		nil,
	)

	s.mcpServer = mcpServer
	s.registerTools()

	return s
}

// MCPServer returns the underlying MCP server instance
func (s *Server) MCPServer() *mcp.Server {
	return s.mcpServer
}

func (s *Server) registerTools() {
	// List releases tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_releases",
		Description: "List all Helm releases in the Kubernetes cluster",
	}, s.handleListReleases)

	// Get release tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_release",
		Description: "Get details of a specific Helm release",
	}, s.handleGetRelease)

	// Get available versions tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_available_versions",
		Description: "Get available chart versions for a Helm release. Requires a registry mapping to be configured for the release.",
	}, s.handleGetAvailableVersions)

	// Upgrade release tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "upgrade_release",
		Description: "Upgrade a Helm release to a specific chart version. Requires a registry mapping to be configured for the release.",
	}, s.handleUpgradeRelease)

	// Get release history tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_release_history",
		Description: "Get the revision history of a Helm release",
	}, s.handleGetReleaseHistory)

	// Get registry mapping tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_registry",
		Description: "Get the registry mapping for a Helm release",
	}, s.handleGetRegistry)

	// Set registry mapping tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "set_registry",
		Description: "Set the registry mapping for a Helm release. This is required before upgrading a release.",
	}, s.handleSetRegistry)

	// Delete registry mapping tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "delete_registry",
		Description: "Delete the registry mapping for a Helm release",
	}, s.handleDeleteRegistry)

	// Get release values tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_release_values",
		Description: "Get the current values (configuration) of a Helm release",
	}, s.handleGetReleaseValues)

	// Update release values tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "update_release_values",
		Description: "Update the values (configuration) of a Helm release. Only the specified values will be updated; existing values are preserved. Requires a registry mapping to be configured for the release.",
	}, s.handleUpdateReleaseValues)
}

func (s *Server) handleListReleases(ctx context.Context, req *mcp.CallToolRequest, input ListReleasesInput) (*mcp.CallToolResult, ListReleasesOutput, error) {
	releases, err := s.helmClient.ListReleases()
	if err != nil {
		return nil, ListReleasesOutput{}, fmt.Errorf("failed to list releases: %w", err)
	}

	// Enrich with registry info and filter
	filteredReleases := make([]model.Release, 0)
	for _, r := range releases {
		// Check registry mapping
		mapping, _ := s.registryStore.GetMapping(ctx, r.Namespace, r.Name)
		r.HasRegistry = mapping != nil

		// Apply namespace filter
		if input.Namespace != "" && r.Namespace != input.Namespace {
			continue
		}

		// Apply hasRegistry filter
		if input.HasRegistry != nil && r.HasRegistry != *input.HasRegistry {
			continue
		}

		filteredReleases = append(filteredReleases, r)
	}

	return nil, ListReleasesOutput{Releases: filteredReleases}, nil
}

func (s *Server) handleGetRelease(ctx context.Context, req *mcp.CallToolRequest, input ReleaseInput) (*mcp.CallToolResult, ReleaseOutput, error) {
	if input.Namespace == "" || input.Name == "" {
		return nil, ReleaseOutput{}, fmt.Errorf("namespace and name are required")
	}

	release, err := s.helmClient.GetRelease(input.Namespace, input.Name)
	if err != nil {
		return nil, ReleaseOutput{}, fmt.Errorf("failed to get release: %w", err)
	}

	// Check registry mapping
	mapping, _ := s.registryStore.GetMapping(ctx, input.Namespace, input.Name)
	release.HasRegistry = mapping != nil

	return nil, ReleaseOutput{Release: release}, nil
}

func (s *Server) handleGetAvailableVersions(ctx context.Context, req *mcp.CallToolRequest, input ReleaseInput) (*mcp.CallToolResult, VersionsOutput, error) {
	if input.Namespace == "" || input.Name == "" {
		return nil, VersionsOutput{}, fmt.Errorf("namespace and name are required")
	}

	versions, err := s.helmClient.GetAvailableVersions(input.Namespace, input.Name)
	if err != nil {
		return nil, VersionsOutput{}, fmt.Errorf("failed to get available versions: %w", err)
	}

	return nil, VersionsOutput{Versions: versions}, nil
}

func (s *Server) handleUpgradeRelease(ctx context.Context, req *mcp.CallToolRequest, input UpgradeInput) (*mcp.CallToolResult, ReleaseOutput, error) {
	if input.Namespace == "" || input.Name == "" || input.ChartVersion == "" {
		return nil, ReleaseOutput{}, fmt.Errorf("namespace, name, and chart_version are required")
	}

	upgradeReq := model.VersionUpgradeRequest{
		ChartVersion: input.ChartVersion,
	}

	release, err := s.helmClient.UpgradeRelease(input.Namespace, input.Name, upgradeReq)
	if err != nil {
		return nil, ReleaseOutput{}, fmt.Errorf("failed to upgrade release: %w", err)
	}

	return nil, ReleaseOutput{Release: release}, nil
}

func (s *Server) handleGetReleaseHistory(ctx context.Context, req *mcp.CallToolRequest, input ReleaseInput) (*mcp.CallToolResult, HistoryOutput, error) {
	if input.Namespace == "" || input.Name == "" {
		return nil, HistoryOutput{}, fmt.Errorf("namespace and name are required")
	}

	history, err := s.helmClient.GetReleaseHistory(input.Namespace, input.Name)
	if err != nil {
		return nil, HistoryOutput{}, fmt.Errorf("failed to get release history: %w", err)
	}

	return nil, HistoryOutput{History: history}, nil
}

func (s *Server) handleGetRegistry(ctx context.Context, req *mcp.CallToolRequest, input RegistryInput) (*mcp.CallToolResult, RegistryOutput, error) {
	if input.Namespace == "" || input.Name == "" {
		return nil, RegistryOutput{}, fmt.Errorf("namespace and name are required")
	}

	mapping, err := s.registryStore.GetMapping(ctx, input.Namespace, input.Name)
	if err != nil {
		return nil, RegistryOutput{}, fmt.Errorf("failed to get registry mapping: %w", err)
	}

	if mapping == nil {
		return nil, RegistryOutput{Message: "no registry mapping configured for this release"}, nil
	}

	return nil, RegistryOutput{Mapping: mapping}, nil
}

func (s *Server) handleSetRegistry(ctx context.Context, req *mcp.CallToolRequest, input SetRegistryInput) (*mcp.CallToolResult, RegistryOutput, error) {
	if input.Namespace == "" || input.Name == "" || input.Registry == "" {
		return nil, RegistryOutput{}, fmt.Errorf("namespace, name, and registry are required")
	}

	// Get release to obtain chart name
	release, err := s.helmClient.GetRelease(input.Namespace, input.Name)
	if err != nil {
		return nil, RegistryOutput{}, fmt.Errorf("failed to get release: %w", err)
	}

	mapping := model.RegistryMapping{
		Namespace:   input.Namespace,
		ReleaseName: input.Name,
		ChartName:   release.Chart,
		Registry:    input.Registry,
	}

	if err := s.registryStore.SetMapping(ctx, mapping); err != nil {
		return nil, RegistryOutput{}, fmt.Errorf("failed to set registry mapping: %w", err)
	}

	return nil, RegistryOutput{Mapping: &mapping}, nil
}

func (s *Server) handleDeleteRegistry(ctx context.Context, req *mcp.CallToolRequest, input RegistryInput) (*mcp.CallToolResult, DeleteRegistryOutput, error) {
	if input.Namespace == "" || input.Name == "" {
		return nil, DeleteRegistryOutput{}, fmt.Errorf("namespace and name are required")
	}

	if err := s.registryStore.DeleteMapping(ctx, input.Namespace, input.Name); err != nil {
		return nil, DeleteRegistryOutput{}, fmt.Errorf("failed to delete registry mapping: %w", err)
	}

	return nil, DeleteRegistryOutput{Success: true, Message: "registry mapping deleted"}, nil
}

func (s *Server) handleGetReleaseValues(ctx context.Context, req *mcp.CallToolRequest, input ReleaseInput) (*mcp.CallToolResult, ValuesOutput, error) {
	if input.Namespace == "" || input.Name == "" {
		return nil, ValuesOutput{}, fmt.Errorf("namespace and name are required")
	}

	values, err := s.helmClient.GetReleaseValues(input.Namespace, input.Name)
	if err != nil {
		return nil, ValuesOutput{}, fmt.Errorf("failed to get release values: %w", err)
	}

	return nil, ValuesOutput{Values: values}, nil
}

func (s *Server) handleUpdateReleaseValues(ctx context.Context, req *mcp.CallToolRequest, input UpdateValuesInput) (*mcp.CallToolResult, ReleaseOutput, error) {
	if input.Namespace == "" || input.Name == "" {
		return nil, ReleaseOutput{}, fmt.Errorf("namespace and name are required")
	}

	if input.Values == nil {
		return nil, ReleaseOutput{}, fmt.Errorf("values are required")
	}

	release, err := s.helmClient.UpdateReleaseValues(input.Namespace, input.Name, input.Values)
	if err != nil {
		return nil, ReleaseOutput{}, fmt.Errorf("failed to update release values: %w", err)
	}

	return nil, ReleaseOutput{Release: release}, nil
}
