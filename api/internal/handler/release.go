package handler

import (
	"net/http"

	"github.com/helm-version-manager/api/internal/helm"
	"github.com/helm-version-manager/api/internal/model"
	"github.com/labstack/echo/v4"
)

type ReleaseHandler struct {
	helmClient *helm.Client
}

func NewReleaseHandler(client *helm.Client) *ReleaseHandler {
	return &ReleaseHandler{
		helmClient: client,
	}
}

func (h *ReleaseHandler) List(c echo.Context) error {
	releases, err := h.helmClient.ListReleases()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, releases)
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
