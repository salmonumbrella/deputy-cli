package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type Location struct {
	Id          int    `json:"Id"`
	CompanyName string `json:"CompanyName"`
	Code        string `json:"Code"`
	Address     string `json:"Address"`
	Active      bool   `json:"Active"`
	Timezone    string `json:"Timezone"`
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
	return locations, err
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
