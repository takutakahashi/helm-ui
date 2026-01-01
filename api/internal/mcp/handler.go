package mcp

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewHTTPHandler creates a new HTTP handler for the MCP server
func (s *Server) NewHTTPHandler() http.Handler {
	logger := slog.Default()

	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			// Return the same server instance for all requests
			return s.mcpServer
		},
		&mcp.StreamableHTTPOptions{
			// Return JSON responses instead of SSE for better client compatibility
			JSONResponse: true,
			Logger:       logger,
		},
	)

	// Wrap with middleware to ensure Accept header compatibility
	// MCP SDK requires both 'application/json' and 'text/event-stream' in Accept header
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("MCP request received",
			"method", r.Method,
			"path", r.URL.Path,
			"accept", r.Header.Get("Accept"),
			"content-type", r.Header.Get("Content-Type"),
		)

		accept := r.Header.Get("Accept")
		hasJSON := strings.Contains(accept, "application/json") || strings.Contains(accept, "application/*") || strings.Contains(accept, "*/*")
		hasSSE := strings.Contains(accept, "text/event-stream") || strings.Contains(accept, "text/*") || strings.Contains(accept, "*/*")

		// If Accept header is missing required content types, add them
		if !hasJSON || !hasSSE {
			r.Header.Set("Accept", "application/json, text/event-stream")
			logger.Info("Accept header modified", "new_accept", "application/json, text/event-stream")
		}

		mcpHandler.ServeHTTP(w, r)
	})
}
