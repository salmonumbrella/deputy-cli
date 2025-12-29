package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type Department struct {
	Id          int    `json:"Id"`
	Company     int    `json:"Company"`
	ParentId    int    `json:"ParentId,omitempty"`
	CompanyName string `json:"CompanyName"`
	CompanyCode string `json:"CompanyCode"`
	Active      bool   `json:"Active"`
	SortOrder   int    `json:"SortOrder,omitempty"`
}

type DepartmentsService struct {
	client *Client
}

func (c *Client) Departments() *DepartmentsService {
	return &DepartmentsService{client: c}
}

func (s *DepartmentsService) List(ctx context.Context, opts *ListOptions) ([]Department, error) {
	var departments []Department

	// Use QUERY endpoint when pagination is specified
	// Simple GET to /resource/OperationalUnit ignores query params
	if opts != nil && (opts.Limit > 0 || opts.Offset > 0) {
		input := &QueryInput{
			Max:   opts.Limit,
			Start: opts.Offset,
		}
		body, err := json.Marshal(input)
		if err != nil {
			return nil, err
		}
		err = s.client.do(ctx, "POST", "/resource/OperationalUnit/QUERY", bytes.NewReader(body), &departments)
		return departments, err
	}

	err := s.client.do(ctx, "GET", "/resource/OperationalUnit", nil, &departments)
	return departments, err
}

func (s *DepartmentsService) Get(ctx context.Context, id int) (*Department, error) {
	var department Department
	path := fmt.Sprintf("/resource/OperationalUnit/%d", id)
	err := s.client.do(ctx, "GET", path, nil, &department)
	return &department, err
}

type CreateDepartmentInput struct {
	Company     int    `json:"intCompanyId"`
	ParentId    int    `json:"intParentId,omitempty"`
	CompanyName string `json:"strCompanyName"`
	CompanyCode string `json:"strCompanyCode,omitempty"`
	SortOrder   int    `json:"intSortOrder,omitempty"`
}

func (s *DepartmentsService) Create(ctx context.Context, input *CreateDepartmentInput) (*Department, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var department Department
	err = s.client.do(ctx, "POST", "/resource/OperationalUnit", bytes.NewReader(body), &department)
	return &department, err
}

type UpdateDepartmentInput struct {
	CompanyName string `json:"strCompanyName,omitempty"`
	CompanyCode string `json:"strCompanyCode,omitempty"`
	Active      *bool  `json:"blnActive,omitempty"`
	SortOrder   int    `json:"intSortOrder,omitempty"`
}

func (s *DepartmentsService) Update(ctx context.Context, id int, input *UpdateDepartmentInput) (*Department, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var department Department
	path := fmt.Sprintf("/resource/OperationalUnit/%d", id)
	err = s.client.do(ctx, "POST", path, bytes.NewReader(body), &department)
	return &department, err
}

func (s *DepartmentsService) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/resource/OperationalUnit/%d", id)
	return s.client.do(ctx, "DELETE", path, nil, nil)
}
