package repository

import (
	"fmt"
	"time"

	"github.com/volatiletech/null/v8"
)

type PublicUser struct {
	Login       string      `json:"login"`
	ID          int         `json:"id"`
	NodeID      string      `json:"node_id"`
	AvatarURL   string      `json:"avatar_url"`
	URL         string      `json:"url"`
	Type        string      `json:"type"`
	Name        null.String `json:"name"`
	Followers   int         `json:"followers"`
	Following   int         `json:"following"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	SuspendedAt null.Time   `json:"suspended_at"`
}

func (r *GitHubClient) GetMe() (*PublicUser, error) {
	if r.authenticatedUser != nil {
		return r.authenticatedUser, nil
	}

	var user PublicUser
	if err := r.restClient.Get("user", &user); err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}
