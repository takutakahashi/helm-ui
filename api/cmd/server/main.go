package main

import (
	"log"
	"net/http"
	"os"

	"github.com/helm-version-manager/api/internal/handler"
	"github.com/helm-version-manager/api/internal/helm"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	helmClient, err := helm.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Helm client: %v", err)
	}

	releaseHandler := handler.NewReleaseHandler(helmClient)
	repoHandler := handler.NewRepositoryHandler(helmClient)

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

	// Repository endpoints
	api.GET("/repositories", repoHandler.List)
	api.POST("/repositories", repoHandler.Add)
	api.DELETE("/repositories/:name", repoHandler.Remove)
	api.POST("/repositories/:name/update", repoHandler.Update)

	// Serve static files (frontend)
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./static"
	}
	if _, err := os.Stat(staticDir); err == nil {
		e.Static("/", staticDir)
		e.File("/", staticDir+"/index.html")
		// SPA fallback
		e.GET("/*", func(c echo.Context) error {
			return c.File(staticDir + "/index.html")
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
