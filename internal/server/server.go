package server

import (
	"net/http"

	stdlibhttp "github.com/andrewhowdencom/stdlib/http"
)

// New creates a new HTTP server with the application's standard configuration.
func New(addr string, handler http.Handler) (*stdlibhttp.Server, error) {
	return stdlibhttp.NewServer(addr, handler)
}
