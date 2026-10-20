package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/unkeyed/unkey/pkg/uid"
)

func TestIndex(t *testing.T) {
	w := request(t, http.MethodGet, "/")
	require.Equal(t, http.StatusOK, w.Code)
	var body indexResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Contains(t, body.Message, "/api/hello")
	require.Contains(t, body.Next, "public until")
	require.Equal(t, "https://www.unkey.com/docs/compute/gateway/api-key-auth", body.Docs)
}

func TestHealth(t *testing.T) {
	w := request(t, http.MethodGet, "/health")
	require.Equal(t, http.StatusOK, w.Code)
	var body healthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, healthResponse{Status: "ok"}, body)
}

func TestHello(t *testing.T) {
	server, err := newServer()
	require.NoError(t, err)
	secret := uid.New(uid.KeyPrefix)
	t.Setenv("DEMO_SECRET", secret)
	r := httptest.NewRequest(http.MethodGet, "/api/hello?key="+secret, nil)
	r.Header.Set("Authorization", "Bearer "+secret)
	r.Header.Set("X-Unkey-Principal", secret)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	require.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var body messageResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, "Hello from Unkey Deploy.", body.Message)
	require.NotEmpty(t, body.RequestID)
	require.NotContains(t, w.Body.String(), secret)
}

func TestRoutingErrors(t *testing.T) {
	for _, tt := range []struct {
		name   string
		method string
		path   string
		status int
		allow  string
	}{
		{name: "unknown", method: http.MethodGet, path: "/missing", status: http.StatusNotFound},
		{name: "unknown API", method: http.MethodGet, path: "/api/missing", status: http.StatusNotFound},
		{name: "wrong method", method: http.MethodPost, path: "/api/hello", status: http.StatusMethodNotAllowed, allow: "GET, HEAD"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := request(t, tt.method, tt.path)
			require.Equal(t, tt.status, w.Code)
			require.Equal(t, tt.allow, w.Header().Get("Allow"))
			var body messageResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
			require.Equal(t, http.StatusText(tt.status), body.Message)
			require.NotEmpty(t, body.RequestID)
		})
	}
}

func TestBodyLimit(t *testing.T) {
	server, err := newServer()
	require.NoError(t, err)
	for _, tt := range []struct {
		name    string
		bytes   int
		status  int
		message string
	}{
		{name: "at limit", bytes: 1024, status: http.StatusOK, message: "Hello from Unkey Deploy."},
		{name: "over limit", bytes: 1025, status: http.StatusRequestEntityTooLarge, message: "Request Entity Too Large"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/hello", strings.NewReader(strings.Repeat("x", tt.bytes)))
			server.ServeHTTP(w, r)
			require.Equal(t, tt.status, w.Code)
			var body messageResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
			require.Equal(t, tt.message, body.Message)
		})
	}
}

func request(t *testing.T, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	server, err := newServer()
	require.NoError(t, err)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, httptest.NewRequest(method, path, nil))
	return w
}
