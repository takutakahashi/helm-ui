package handler

import (
	"net/http"

	"github.com/helm-version-manager/api/internal/helm"
	"github.com/helm-version-manager/api/internal/model"
	"github.com/labstack/echo/v4"
)

type RepositoryHandler struct {
	helmClient *helm.Client
}

func NewRepositoryHandler(client *helm.Client) *RepositoryHandler {
	return &RepositoryHandler{
		helmClient: client,
	}
}

func (h *RepositoryHandler) List(c echo.Context) error {
	repos, err := h.helmClient.ListRepositories()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, repos)
}

func (h *RepositoryHandler) Add(c echo.Context) error {
	var req model.AddRepositoryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.Name == "" || req.URL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name and url are required")
	}

	if err := h.helmClient.AddRepository(req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Repository added successfully"})
}

func (h *RepositoryHandler) Remove(c echo.Context) error {
	name := c.Param("name")

	if err := h.helmClient.RemoveRepository(name); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *RepositoryHandler) Update(c echo.Context) error {
	name := c.Param("name")

	if err := h.helmClient.UpdateRepository(name); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Repository updated successfully"})
}
