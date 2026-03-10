// Package traefik_requestid is a Traefik plugin that injects a unique
// request ID into incoming requests if one is not already present.
package traefik_requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// Config holds the plugin configuration.
type Config struct {
	// HeaderName is the header used to read/write the request ID.
	// Defaults to "X-Request-ID".
	HeaderName string `json:"headerName,omitempty"`
	// OverrideExisting forces a new ID even if the header is already set.
	OverrideExisting bool `json:"overrideExisting,omitempty"`
	// ForwardToResponse mirrors the request ID onto the response headers.
	ForwardToResponse bool `json:"forwardToResponse,omitempty"`
}

// CreateConfig returns a Config with sensible defaults.
func CreateConfig() *Config {
	return &Config{
		HeaderName:        "X-Request-ID",
		OverrideExisting:  false,
		ForwardToResponse: true,
	}
}

// RequestID is the middleware handler.
type RequestID struct {
	next              http.Handler
	headerName        string
	overrideExisting  bool
	forwardToResponse bool
	name              string
}

// New creates a new RequestID middleware instance.
func New(_ context.Context, next http.Handler, cfg *Config, name string) (http.Handler, error) {
	headerName := cfg.HeaderName
	if headerName == "" {
		headerName = "X-Request-ID"
	}
	return &RequestID{
		next:              next,
		headerName:        headerName,
		overrideExisting:  cfg.OverrideExisting,
		forwardToResponse: cfg.ForwardToResponse,
		name:              name,
	}, nil
}

// ServeHTTP implements http.Handler.
func (r *RequestID) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	id := req.Header.Get(r.headerName)

	if id == "" || r.overrideExisting {
		id = generateID()
		req.Header.Set(r.headerName, id)
	}

	if r.forwardToResponse {
		rw.Header().Set(r.headerName, id)
	}

	r.next.ServeHTTP(rw, req)
}

// generateID returns a UUID v4 string without dashes (32 hex chars).
func generateID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "00000000000000000000000000000000"
	}
	// Set UUID v4 version bits (0100xxxx) and variant bits (10xxxxxx)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return hex.EncodeToString(b)
}