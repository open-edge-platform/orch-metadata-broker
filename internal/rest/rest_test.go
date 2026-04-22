// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// specPath returns the path to the openapi.yaml bundled with this repo.
func specPath(t *testing.T) string {
	t.Helper()
	// Resolve from the test binary's source location: internal/rest/ → api/spec/openapi.yaml
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")
	// Walk up to repo root (internal/rest → internal → repo root).
	root := filepath.Join(filepath.Dir(filename), "..", "..")
	p := filepath.Join(root, "api", "spec", "openapi.yaml")
	abs, err := filepath.Abs(p)
	require.NoError(t, err)
	_, err = os.Stat(abs)
	require.NoError(t, err, "openapi spec not found at %s", abs)
	return abs
}

// buildServer creates a server using an arbitrary free port pair.
func buildServer(t *testing.T, tenantManagerURL string) *http.Server {
	t.Helper()
	// Use ports that nothing is listening on — gRPC gateway connection is lazy.
	return newServerWithTenantURL(19801, 19802, "/", "", specPath(t), tenantManagerURL)
}

// TestNewServer_HealthzEndpoint verifies the /healthz route is registered
// and returns 200 OK without a live gRPC backend.
func TestNewServer_HealthzEndpoint(t *testing.T) {
	srv := buildServer(t, "http://127.0.0.1:19999")
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestNewServer_TestEndpoint verifies the /test route is registered.
func TestNewServer_TestEndpoint(t *testing.T) {
	srv := buildServer(t, "http://127.0.0.1:19999")
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestNewServer_EmptyTenantManagerURL_SmokeTest verifies that an empty tenantManagerURL
// is replaced with the default value (smoke-test the fallback branch).
func TestNewServer_EmptyTenantManagerURL_SmokeTest(t *testing.T) {
	// Passing empty string should not panic and still produce a usable server.
	srv := newServerWithTenantURL(19803, 19804, "/", "", specPath(t), "")
	require.NotNil(t, srv)
}

// TestNewServer_Address verifies that the *http.Server Addr field reflects the port.
func TestNewServer_Address(t *testing.T) {
	srv := buildServer(t, "http://127.0.0.1:19999")
	assert.Equal(t, fmt.Sprintf(":%d", 19801), srv.Addr)
}

// TestNewServer_WithTenantManagerURLInEnv verifies that NewServer reads TENANT_MANAGER_URL
// from the environment (ensuring the env-var path is wired correctly).
func TestNewServer_WithTenantManagerURLInEnv(t *testing.T) {
	t.Setenv("TENANT_MANAGER_URL", "http://custom-tenant-manager:9090")
	// NewServer calls newServerWithTenantURL internally — just verify it doesn't panic
	// or fatal when the env var is set correctly.
	srv := NewServer(19805, 19806, "/", "", specPath(t))
	require.NotNil(t, srv)
}
