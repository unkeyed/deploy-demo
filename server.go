package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/unkeyed/unkey/pkg/codes"
	"github.com/unkeyed/unkey/pkg/fault"
	"github.com/unkeyed/unkey/pkg/zen"
)

type indexResponse struct {
	Message string `json:"message"`
	Next    string `json:"next"`
	Docs    string `json:"docs"`
}

type healthResponse struct {
	Status string `json:"status"`
}

type messageResponse struct {
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

func newServer() (*zen.Server, error) {
	server, err := zen.New(zen.Config{MaxRequestBodySize: 1024})
	if err != nil {
		return nil, err
	}

	middleware := []zen.Middleware{zen.WithPanicRecovery(), withErrors}
	server.RegisterRoutingErrors(middleware)
	server.RegisterRoute(middleware, zen.NewRoute(http.MethodGet, "/{$}", index))
	server.RegisterRoute(middleware, zen.NewRoute(http.MethodGet, "/health", health))
	server.RegisterRoute(middleware, zen.NewRoute(http.MethodGet, "/api/hello", hello))
	server.RegisterRoute(middleware, zen.NewRoute(http.MethodGet, "/api/whoami", whoami))
	return server, nil
}

func index(_ context.Context, s *zen.Session) error {
	return s.JSON(http.StatusOK, indexResponse{
		Message: "Your Unkey deployment is running. Try GET /api/hello or GET /api/whoami.",
		Next:    "The greeting is public until you add a Key Auth policy for /api/ and redeploy.",
		Docs:    "https://www.unkey.com/docs/compute/gateway/api-key-auth",
	})
}

func health(_ context.Context, s *zen.Session) error {
	return s.JSON(http.StatusOK, healthResponse{Status: "ok"})
}

func hello(_ context.Context, s *zen.Session) error {
	return s.JSON(http.StatusOK, messageResponse{
		Message:   "Hello from Unkey Deploy.",
		RequestID: s.RequestID(),
	})
}

func withErrors(next zen.HandleFunc) zen.HandleFunc {
	return func(ctx context.Context, s *zen.Session) error {
		s.ResponseWriter().Header().Set("Cache-Control", "no-store")
		err := next(ctx, s)
		if err == nil {
			return nil
		}

		status := http.StatusInternalServerError
		var routingError *zen.RoutingError
		if errors.As(err, &routingError) {
			status = routingError.Status
		} else if code, ok := fault.GetCode(err); ok && code == codes.UserErrorsBadRequestRequestBodyTooLarge {
			status = http.StatusRequestEntityTooLarge
		}
		return s.JSON(status, messageResponse{
			Message:   http.StatusText(status),
			RequestID: s.RequestID(),
		})
	}
}
