package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/helm-version-manager/api/internal/handler"
	"github.com/helm-version-manager/api/internal/helm"
	mcpserver "github.com/helm-version-manager/api/internal/mcp"
	"github.com/helm-version-manager/api/internal/storage"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	registryStore, err := storage.NewRegistryStore()
	if err != nil {
		log.Fatalf("Failed to create registry store: %v", err)
	}

	helmClient, err := helm.NewClient(registryStore)
	if err != nil {
		log.Fatalf("Failed to create Helm client: %v", err)
	}

	releaseHandler := handler.NewReleaseHandler(helmClient, registryStore)

	// Create MCP server
	mcpServer := mcpserver.NewServer(helmClient, registryStore)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	api := e.Group("/api")

	// Release endpoints
	api.GET("/releases", releaseHandler.List)
	api.GET("/releases/:namespace/:name", releaseHandler.Get)
	api.GET("/releases/:namespace/:name/versions", releaseHandler.GetVersions)
	api.PUT("/releases/:namespace/:name", releaseHandler.Upgrade)
	api.GET("/releases/:namespace/:name/history", releaseHandler.GetHistory)

	// Registry mapping endpoints
	api.GET("/releases/:namespace/:name/registry", releaseHandler.GetRegistry)
	api.PUT("/releases/:namespace/:name/registry", releaseHandler.SetRegistry)
	api.DELETE("/releases/:namespace/:name/registry", releaseHandler.DeleteRegistry)

	// MCP server endpoint (Streamable HTTP)
	mcpHandler := echo.WrapHandler(mcpServer.NewHTTPHandler())
	e.Any("/mcp", mcpHandler)
	e.Any("/mcp/*", mcpHandler)

	// Serve static files (frontend)
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./static"
	}
	if _, err := os.Stat(staticDir); err == nil {
		// Serve static files with SPA fallback
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				path := c.Request().URL.Path
				// Skip API, health, and MCP routes (don't apply SPA fallback)
				if strings.HasPrefix(path, "/api") ||
					strings.HasPrefix(path, "/mcp") ||
					strings.HasPrefix(path, "/.well-known") ||
					path == "/health" {
					return next(c)
				}
				// Try to serve static file
				filePath := staticDir + path
				if _, err := os.Stat(filePath); err == nil {
					return c.File(filePath)
				}
				// SPA fallback: serve index.html for non-existent paths
				return c.File(staticDir + "/index.html")
			}
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
