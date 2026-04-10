package v1_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/itunified-io/dbx/api/v1"
	"github.com/itunified-io/dbx/pkg/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
	s := v1.NewServer(pipeline.New())
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "ok", body["status"])
}

func TestVersionEndpoint(t *testing.T) {
	s := v1.NewServer(pipeline.New())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	w := httptest.NewRecorder()

	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "dev", body["version"])
}

func TestTargetsEndpoint(t *testing.T) {
	s := v1.NewServer(pipeline.New())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/targets", nil)
	w := httptest.NewRecorder()

	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
