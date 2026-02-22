package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type SalesData struct {
	Id        int     `json:"Id"`
	Company   int     `json:"Company"`
	Area      int     `json:"Area,omitempty"`
	Timestamp int64   `json:"Timestamp"`
	Value     float64 `json:"Value"`
	Type      string  `json:"Type,omitempty"`
}

type SalesService struct {
	client *Client
}

func (c *Client) Sales() *SalesService {
	return &SalesService{client: c}
}

func (s *SalesService) List(ctx context.Context) ([]SalesData, error) {
	var sales []SalesData
	err := s.client.do(ctx, "GET", "/resource/SalesData", nil, &sales)
	return sales, err
}

type CreateSalesInput struct {
	Company   int     `json:"intCompanyId"`
	Area      int     `json:"intAreaId,omitempty"`
	Timestamp int64   `json:"intTimestamp"`
	Value     float64 `json:"fltValue"`
	Type      string  `json:"strType,omitempty"`
}

// Add uses the v2 metrics API
func (s *SalesService) Add(ctx context.Context, input *CreateSalesInput) (*SalesData, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var sale SalesData
	// Note: v2 API endpoint
	err = s.client.doV2(ctx, "POST", "/metrics", bytes.NewReader(body), &sale)
	return &sale, err
}

type SalesQueryInput struct {
	Company   int   `json:"intCompanyId,omitempty"`
	StartTime int64 `json:"intStartTime,omitempty"`
	EndTime   int64 `json:"intEndTime,omitempty"`
}

func (s *SalesService) Query(ctx context.Context, input *SalesQueryInput) ([]SalesData, error) {
	path := fmt.Sprintf("/resource/SalesData?company=%d", input.Company)
	if input.StartTime > 0 {
		path += fmt.Sprintf("&start=%d", input.StartTime)
	}
	if input.EndTime > 0 {
		path += fmt.Sprintf("&end=%d", input.EndTime)
	}

	var sales []SalesData
	err := s.client.do(ctx, "GET", path, nil, &sales)
	return sales, err
}
