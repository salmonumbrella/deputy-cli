package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type ResourceInfo struct {
	Name   string                 `json:"name"`
	Fields map[string]interface{} `json:"fields"`
	Assocs interface{}            `json:"assocs,omitempty"`
}

// AssocsAsMap returns associations as a map if the API returned a map,
// or nil if associations are in array format or empty.
func (r *ResourceInfo) AssocsAsMap() map[string]interface{} {
	if r.Assocs == nil {
		return nil
	}
	if m, ok := r.Assocs.(map[string]interface{}); ok {
		return m
	}
	return nil
}

// AssocsAsArray returns associations as a slice of strings if the API returned an array,
// or nil if associations are in map format or empty.
func (r *ResourceInfo) AssocsAsArray() []string {
	if r.Assocs == nil {
		return nil
	}
	arr, ok := r.Assocs.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// HasAssocs returns true if there are any associations defined.
func (r *ResourceInfo) HasAssocs() bool {
	if r.Assocs == nil {
		return false
	}
	if m := r.AssocsAsMap(); m != nil {
		return len(m) > 0
	}
	if a := r.AssocsAsArray(); a != nil {
		return len(a) > 0
	}
	return false
}

type ResourceService struct {
	client       *Client
	resourceName string
}

func (c *Client) Resource(name string) *ResourceService {
	return &ResourceService{client: c, resourceName: name}
}

func (s *ResourceService) Info(ctx context.Context) (*ResourceInfo, error) {
	var info ResourceInfo
	path := fmt.Sprintf("/resource/%s/INFO", s.resourceName)
	err := s.client.do(ctx, "GET", path, nil, &info)
	return &info, err
}

type QueryInput struct {
	Search map[string]interface{} `json:"search,omitempty"`
	Join   []string               `json:"join,omitempty"`
	Assoc  []string               `json:"assoc,omitempty"`
	Sort   map[string]string      `json:"sort,omitempty"`
	Max    int                    `json:"max,omitempty"`
	Start  int                    `json:"start,omitempty"`
}

func (s *ResourceService) Query(ctx context.Context, input *QueryInput) ([]map[string]interface{}, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	path := fmt.Sprintf("/resource/%s/QUERY", s.resourceName)
	err = s.client.do(ctx, "POST", path, bytes.NewReader(body), &results)
	return results, err
}

func (s *ResourceService) Get(ctx context.Context, id int) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/resource/%s/%d", s.resourceName, id)
	err := s.client.do(ctx, "GET", path, nil, &result)
	return result, err
}

func (s *ResourceService) List(ctx context.Context) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	path := fmt.Sprintf("/resource/%s", s.resourceName)
	err := s.client.do(ctx, "GET", path, nil, &results)
	return results, err
}

// KnownResources returns a list of common Deputy resource names
func KnownResources() []string {
	return []string{
		"Employee",
		"EmployeeRole",
		"Company",
		"OperationalUnit",
		"Timesheet",
		"PayRules",
		"TimesheetPayReturn",
		"Roster",
		"Leave",
		"LeaveRules",
		"LeaveAccrualTransaction",
		"Address",
		"Contact",
		"EmploymentContract",
		"EmployeeAvailability",
		"EmployeeSalaryOpunitCosting",
		"EmployeeAppraisal",
		"EmployeeAgreement",
		"EmployeeHistory",
		"TrainingModule",
		"TrainingRecord",
		"Task",
		"Memo",
		"Journal",
		"Comment",
		"Webhook",
		"SalesData",
		"SystemUsageTracking",
	}
}
