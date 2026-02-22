package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
)

type Location struct {
	Id          int             `json:"Id"`
	CompanyName string          `json:"CompanyName"`
	Code        string          `json:"Code"`
	CompanyCode string          `json:"CompanyCode,omitempty"`
	Address     json.RawMessage `json:"Address"` // Can be string or int (foreign key)
	Active      bool            `json:"Active"`
	Timezone    string          `json:"Timezone"`
}

// AddressString returns Address as a string for display purposes.
// When Address is an integer (foreign key to Address table), returns "(ref:N)".
func (l *Location) AddressString() string {
	if len(l.Address) == 0 || string(l.Address) == "null" {
		return ""
	}
	// Try string first
	var s string
	if err := json.Unmarshal(l.Address, &s); err == nil {
		return s
	}
	// Try number (foreign key)
	var n float64
	if err := json.Unmarshal(l.Address, &n); err == nil {
		return fmt.Sprintf("(ref:%d)", int(n))
	}
	return string(l.Address)
}

type LocationsService struct {
	client *Client
}

func (c *Client) Locations() *LocationsService {
	return &LocationsService{client: c}
}

func (s *LocationsService) List(ctx context.Context, opts *ListOptions) ([]Location, error) {
	var locations []Location
	err := s.client.doWithOpts(ctx, "GET", "/supervise/location/simplified", nil, &locations, opts)
	if err == nil {
		return locations, nil
	}

	if !IsNotFound(err) && !IsForbidden(err) {
		return nil, err
	}

	if s.client.debug {
		_, _ = fmt.Fprintln(os.Stderr, "Debug: locations list fallback to /resource/Company")
	}

	var fallback []Location
	fallbackErr := s.client.doWithOpts(ctx, "GET", "/resource/Company", nil, &fallback, opts)
	if fallbackErr == nil {
		return fallback, nil
	}

	return nil, fmt.Errorf("locations list failed: %w (fallback to /resource/Company failed: %v)", err, fallbackErr)
}

func (s *LocationsService) Get(ctx context.Context, id int) (*Location, error) {
	var location Location
	path := fmt.Sprintf("/resource/Company/%d", id)
	err := s.client.do(ctx, "GET", path, nil, &location)
	return &location, err
}

type CreateLocationInput struct {
	CompanyName string `json:"strCompanyName"`
	Code        string `json:"strCompanyCode,omitempty"`
	Address     string `json:"strAddress,omitempty"`
	Timezone    string `json:"strTimezone,omitempty"`
}

func (s *LocationsService) Create(ctx context.Context, input *CreateLocationInput) (*Location, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var location Location
	err = s.client.do(ctx, "POST", "/supervise/location", bytes.NewReader(body), &location)
	return &location, err
}

type UpdateLocationInput struct {
	CompanyName string `json:"strCompanyName,omitempty"`
	Code        string `json:"strCompanyCode,omitempty"`
	Address     string `json:"strAddress,omitempty"`
	Timezone    string `json:"strTimezone,omitempty"`
}

func (s *LocationsService) Update(ctx context.Context, id int, input *UpdateLocationInput) (*Location, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var location Location
	path := fmt.Sprintf("/supervise/location/%d", id)
	err = s.client.do(ctx, "PUT", path, bytes.NewReader(body), &location)
	return &location, err
}

func (s *LocationsService) Archive(ctx context.Context, id int) error {
	path := fmt.Sprintf("/supervise/location/%d/archive", id)
	return s.client.do(ctx, "POST", path, nil, nil)
}

func (s *LocationsService) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/supervise/location/%d", id)
	return s.client.do(ctx, "DELETE", path, nil, nil)
}

type LocationSettings struct {
	Id       int                    `json:"Id"`
	Settings map[string]interface{} `json:"Settings"`
}

func (s *LocationsService) GetSettings(ctx context.Context, id int) (*LocationSettings, error) {
	var settings LocationSettings
	path := fmt.Sprintf("/supervise/location/%d/settings", id)
	err := s.client.do(ctx, "GET", path, nil, &settings)
	return &settings, err
}

type UpdateSettingsInput struct {
	Settings map[string]interface{} `json:"arrSettings"`
}

func (s *LocationsService) UpdateSettings(ctx context.Context, id int, settings map[string]interface{}) error {
	input := UpdateSettingsInput{Settings: settings}
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/supervise/location/%d/settings", id)
	return s.client.do(ctx, "POST", path, bytes.NewReader(body), nil)
}
