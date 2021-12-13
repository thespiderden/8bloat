package mastodon

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type NotificationPleroma struct {
	IsSeen bool `json:"is_seen"`
}

// Notification hold information for mastodon notification.
type Notification struct {
	ID        string               `json:"id"`
	Type      string               `json:"type"`
	CreatedAt time.Time            `json:"created_at"`
	Account   Account              `json:"account"`
	Status    *Status              `json:"status"`
	Pleroma   *NotificationPleroma `json:"pleroma"`
}

// GetNotifications return notifications.
func (c *Client) GetNotifications(ctx context.Context, pg *Pagination, includes, excludes []string) ([]*Notification, error) {
	var notifications []*Notification
	params := url.Values{}
	for _, include := range includes {
		params.Add("include_types[]", include)
	}
	for _, exclude := range excludes {
		params.Add("exclude_types[]", exclude)
	}
	err := c.doAPI(ctx, http.MethodGet, "/api/v1/notifications", params, &notifications, pg)
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// GetNotification return notification.
func (c *Client) GetNotification(ctx context.Context, id string) (*Notification, error) {
	var notification Notification
	err := c.doAPI(ctx, http.MethodGet, fmt.Sprintf("/api/v1/notifications/%v", id), nil, &notification, nil)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// ClearNotifications clear notifications.
func (c *Client) ClearNotifications(ctx context.Context) error {
	return c.doAPI(ctx, http.MethodPost, "/api/v1/notifications/clear", nil, nil, nil)
}

// ReadNotifications marks notifications as read
// Currenly only works for Pleroma
func (c *Client) ReadNotifications(ctx context.Context, maxID string) error {
	params := url.Values{}
	params.Set("max_id", maxID)
	return c.doAPI(ctx, http.MethodPost, "/api/v1/pleroma/notifications/read", params, nil, nil)
}
