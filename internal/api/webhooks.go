package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type Webhook struct {
	Id       int    `json:"Id"`
	Topic    string `json:"Topic"`
	Url      string `json:"Address"`
	Type     string `json:"Type,omitempty"`
	Enabled  bool   `json:"Enabled"`
	Created  string `json:"Created,omitempty"`
	Modified string `json:"Modified,omitempty"`
}

type WebhooksService struct {
	client *Client
}

func (c *Client) Webhooks() *WebhooksService {
	return &WebhooksService{client: c}
}

func (s *WebhooksService) List(ctx context.Context, opts *ListOptions) ([]Webhook, error) {
	var webhooks []Webhook
	err := s.client.doWithOpts(ctx, "GET", "/resource/Webhook", nil, &webhooks, opts)
	return webhooks, err
}

func (s *WebhooksService) Get(ctx context.Context, id int) (*Webhook, error) {
	var webhook Webhook
	path := fmt.Sprintf("/resource/Webhook/%d", id)
	err := s.client.do(ctx, "GET", path, nil, &webhook)
	return &webhook, err
}

type CreateWebhookInput struct {
	Topic   string `json:"strTopic"`
	Url     string `json:"strUrl"`
	Type    string `json:"strType,omitempty"`
	Enabled bool   `json:"blnEnabled"`
}

func (s *WebhooksService) Create(ctx context.Context, input *CreateWebhookInput) (*Webhook, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var webhook Webhook
	err = s.client.do(ctx, "POST", "/resource/Webhook", bytes.NewReader(body), &webhook)
	return &webhook, err
}

func (s *WebhooksService) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/resource/Webhook/%d", id)
	return s.client.do(ctx, "DELETE", path, nil, nil)
}
