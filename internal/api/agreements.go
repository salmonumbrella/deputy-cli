package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type EmployeeAgreement struct {
	Id       int             `json:"Id"`
	Employee int             `json:"Employee"`
	Active   bool            `json:"Active"`
	BaseRate *float64        `json:"BaseRate,omitempty"`
	Config   json.RawMessage `json:"Config,omitempty"`
	Contract int             `json:"Contract,omitempty"`
	PayPoint int             `json:"PayPoint,omitempty"`
}

type AgreementsService struct {
	client *Client
}

func (c *Client) Agreements() *AgreementsService {
	return &AgreementsService{client: c}
}

// ListByEmployee returns agreements for a specific employee.
func (s *AgreementsService) ListByEmployee(ctx context.Context, employeeID int, activeOnly bool) ([]EmployeeAgreement, error) {
	search := map[string]interface{}{
		"s1": map[string]interface{}{"field": "EmployeeId", "type": "eq", "data": employeeID},
	}
	if activeOnly {
		search["s2"] = map[string]interface{}{"field": "Active", "type": "eq", "data": true}
	}

	input := &QueryInput{Search: search}
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var agreements []EmployeeAgreement
	err = s.client.do(ctx, "POST", "/resource/EmployeeAgreement/QUERY", bytes.NewReader(body), &agreements)
	return agreements, err
}

// Get returns a specific agreement by ID.
func (s *AgreementsService) Get(ctx context.Context, id int) (*EmployeeAgreement, error) {
	var agreement EmployeeAgreement
	path := fmt.Sprintf("/resource/EmployeeAgreement/%d", id)
	err := s.client.do(ctx, "GET", path, nil, &agreement)
	return &agreement, err
}

// UpdateAgreementInput updates selected agreement fields.
type UpdateAgreementInput struct {
	BaseRate *float64         `json:"BaseRate,omitempty"`
	Config   *json.RawMessage `json:"Config,omitempty"`
}

// Update updates an agreement by ID.
func (s *AgreementsService) Update(ctx context.Context, id int, input *UpdateAgreementInput) (*EmployeeAgreement, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var agreement EmployeeAgreement
	path := fmt.Sprintf("/resource/EmployeeAgreement/%d", id)
	err = s.client.do(ctx, "POST", path, bytes.NewReader(body), &agreement)
	return &agreement, err
}
