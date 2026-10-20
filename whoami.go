package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/unkeyed/unkey/pkg/zen"
)

type gatewayPrincipal struct {
	Version string `json:"version"`
	Type    string `json:"type"`
	Subject string `json:"subject"`
	Source  struct {
		Key gatewayKey `json:"key"`
	} `json:"source"`
}

type gatewayKey struct {
	KeyID       string   `json:"keyId"`
	KeySpaceID  string   `json:"keySpaceId"`
	Permissions []string `json:"permissions"`
}

type whoamiResponse struct {
	Subject     string   `json:"subject"`
	KeyID       string   `json:"keyId"`
	KeySpaceID  string   `json:"keySpaceId"`
	Permissions []string `json:"permissions,omitempty"`
	RequestID   string   `json:"requestId"`
}

func whoami(_ context.Context, s *zen.Session) error {
	raw := s.Request().Header.Get("X-Unkey-Principal")
	if raw == "" {
		return s.JSON(http.StatusUnauthorized, messageResponse{
			Message:   "No gateway principal. Add a Key Auth policy for /api/ and redeploy.",
			RequestID: s.RequestID(),
		})
	}

	invalid := messageResponse{Message: "Invalid gateway principal.", RequestID: s.RequestID()}
	var principal gatewayPrincipal
	if err := json.Unmarshal([]byte(raw), &principal); err != nil {
		return s.JSON(http.StatusBadGateway, invalid)
	}
	if principal.Version != "v1" || principal.Type != "API_KEY" {
		return s.JSON(http.StatusBadGateway, invalid)
	}
	key := principal.Source.Key
	if principal.Subject == "" || key.KeyID == "" || key.KeySpaceID == "" {
		return s.JSON(http.StatusBadGateway, invalid)
	}

	return s.JSON(http.StatusOK, whoamiResponse{
		Subject:     principal.Subject,
		KeyID:       key.KeyID,
		KeySpaceID:  key.KeySpaceID,
		Permissions: key.Permissions,
		RequestID:   s.RequestID(),
	})
}
