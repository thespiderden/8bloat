package masta

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Filter struct {
	ID           string     `json:"id"`
	Phrase       string     `json:"phrase"`
	Context      []string   `json:"context"`
	WholeWord    bool       `json:"whole_word"`
	ExpiresAt    *time.Time `json:"expires_at"`
	Irreversible bool       `json:"irreversible"`
}

func (c *Client) GetFilters(ctx context.Context) ([]*Filter, error) {
	var filters []*Filter
	err := c.doAPI(ctx, http.MethodGet, "/api/v1/filters", nil, &filters, nil)
	if err != nil {
		return nil, err
	}
	return filters, nil
}

func (c *Client) AddFilter(ctx context.Context, phrase string, context []string, irreversible bool, wholeWord bool, expiresIn *time.Time) error {
	params := url.Values{}
	params.Set("phrase", phrase)
	for i := range context {
		params.Add("context[]", context[i])
	}
	params.Set("irreversible", strconv.FormatBool(irreversible))
	params.Set("whole_word", strconv.FormatBool(wholeWord))
	if expiresIn != nil {
		params.Set("expires_in", expiresIn.Format(time.RFC3339))
	}
	err := c.doAPI(ctx, http.MethodPost, "/api/v1/filters", params, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) RemoveFilter(ctx context.Context, id string) error {
	return c.doAPI(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/filters/%s", id), nil, nil, nil)
}
