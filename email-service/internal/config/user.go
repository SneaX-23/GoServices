package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Base URL for user service
const UserServiceURL = "http://user-service:3000"

// Struct for decoding response
type UserEmail struct {
	Email string `json:"email"`
}

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

// fetches the email address for a given user ID
func GetUserEmail(ctx context.Context, userID string) (string, error) {
	url := fmt.Sprintf("%s/api/users/%s", UserServiceURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		slog.Error("failed to create request", "err", err)
		return "", err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Error("failed to call user service", "err", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn(
			"user service returned non-200",
			"status", resp.StatusCode,
			"user_id", userID,
		)
		return "", errors.New("user service returned non-OK status")
	}

	var res UserEmail
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		slog.Error("failed to decode response", "err", err)
		return "", err
	}

	return res.Email, nil
}
