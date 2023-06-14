package masta

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Poll struct {
	ID          string       `json:"id"`
	ExpiresAt   *time.Time    `json:"expires_at"`
	Expired     bool         `json:"expired"`
	Multiple    bool         `json:"multiple"`
	VotesCount  int64        `json:"votes_count"`
	Voted       bool         `json:"voted"`
	Emojis      []Emoji      `json:"emojis"`
	Options     []PollOption `json:"options"`
}

// Poll hold information for a masta poll option.
type PollOption struct {
	Title      string `json:"title"`
	VotesCount int64  `json:"votes_count"`
}

// Vote submits a vote with given choices to the poll specified by id.
func (c *Client) Vote(ctx context.Context, id string, choices []string) (*Poll, error) {
	var poll Poll
	params := make(url.Values)
	params["choices[]"] = choices
	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/polls/%s/votes", id), params, &poll, nil)
	if err != nil {
		return nil, err
	}
	return &poll, nil
}
