package mcp

import (
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewHTTPHandler creates a new HTTP handler for the MCP server
func (s *Server) NewHTTPHandler() http.Handler {
	return mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			// Return the same server instance for all requests
			return s.mcpServer
		},
		nil,
	)
}
