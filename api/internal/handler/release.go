package handler

import (
	"net/http"

	"github.com/helm-version-manager/api/internal/helm"
	"github.com/helm-version-manager/api/internal/model"
	"github.com/helm-version-manager/api/internal/storage"
	"github.com/labstack/echo/v4"
)

type ReleaseHandler struct {
	helmClient    *helm.Client
	registryStore *storage.RegistryStore
}

func NewReleaseHandler(client *helm.Client, store *storage.RegistryStore) *ReleaseHandler {
	return &ReleaseHandler{
		helmClient:    client,
		registryStore: store,
	}
}

func (h *ReleaseHandler) List(c echo.Context) error {
	releases, err := h.helmClient.ListReleases()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Get all registry mappings
	mappings, err := h.registryStore.ListMappings(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Create a set of releases that have registry mappings
	registrySet := make(map[string]bool)
	for _, m := range mappings {
		key := m.Namespace + "/" + m.ReleaseName
		registrySet[key] = true
	}

	// Mark releases with registry mappings
	for i := range releases {
		key := releases[i].Namespace + "/" + releases[i].Name
		releases[i].HasRegistry = registrySet[key]
	}

	// Apply filters
	namespaceFilter := c.QueryParam("namespace")
	hasRegistryFilter := c.QueryParam("hasRegistry")

	var filteredReleases []model.Release
	for _, r := range releases {
		// Filter by namespace
		if namespaceFilter != "" && r.Namespace != namespaceFilter {
			continue
		}

		// Filter by hasRegistry
		if hasRegistryFilter != "" {
			hasReg := hasRegistryFilter == "true"
			if r.HasRegistry != hasReg {
				continue
			}
		}

		filteredReleases = append(filteredReleases, r)
	}

	if filteredReleases == nil {
		filteredReleases = []model.Release{}
	}

	return c.JSON(http.StatusOK, filteredReleases)
}

func (h *ReleaseHandler) Get(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")

	release, err := h.helmClient.GetRelease(namespace, name)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, release)
}

func (h *ReleaseHandler) GetVersions(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")

	versions, err := h.helmClient.GetAvailableVersions(namespace, name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, versions)
}

func (h *ReleaseHandler) Upgrade(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var req model.VersionUpgradeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.ChartVersion == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "chartVersion is required")
	}

	release, err := h.helmClient.UpgradeRelease(namespace, name, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, release)
}

func (h *ReleaseHandler) GetHistory(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")

	history, err := h.helmClient.GetReleaseHistory(namespace, name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, history)
}

func (h *ReleaseHandler) GetRegistry(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")

	mapping, err := h.registryStore.GetMapping(c.Request().Context(), namespace, name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if mapping == nil {
		return echo.NewHTTPError(http.StatusNotFound, "registry mapping not found")
	}

	return c.JSON(http.StatusOK, mapping)
}

func (h *ReleaseHandler) SetRegistry(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var req model.SetRegistryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.Registry == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "registry is required")
	}

	release, err := h.helmClient.GetRelease(namespace, name)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	mapping := model.RegistryMapping{
		Namespace:   namespace,
		ReleaseName: name,
		ChartName:   release.Chart,
		Registry:    req.Registry,
	}

	if err := h.registryStore.SetMapping(c.Request().Context(), mapping); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mapping)
}

func (h *ReleaseHandler) DeleteRegistry(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if err := h.registryStore.DeleteMapping(c.Request().Context(), namespace, name); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ReleaseHandler) GetValues(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")

	values, err := h.helmClient.GetReleaseValues(namespace, name)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, values)
}

func (h *ReleaseHandler) UpdateValues(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var req model.ValuesUpdateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.Values == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "values is required")
	}

	release, err := h.helmClient.UpdateReleaseValues(namespace, name, req.Values)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, release)
}

func (h *ReleaseHandler) Rollback(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var req model.RollbackRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.Revision <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "revision must be a positive integer")
	}

	release, err := h.helmClient.RollbackRelease(namespace, name, req.Revision)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, release)
}
