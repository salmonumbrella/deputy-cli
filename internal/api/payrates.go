package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type PayRatesService struct {
	client *Client
}

func (c *Client) PayRates() *PayRatesService {
	return &PayRatesService{client: c}
}

// AwardLibraryEntry is a flexible container for award library data.
type AwardLibraryEntry map[string]interface{}

// ListAwardsLibrary returns the award library entries available to the install.
func (s *PayRatesService) ListAwardsLibrary(ctx context.Context) ([]AwardLibraryEntry, error) {
	var awards []AwardLibraryEntry
	err := s.client.do(ctx, "GET", "/payroll/listAwardsLibrary", nil, &awards)
	return awards, err
}

// GetAwardFromLibrary returns a specific award from the award library by code.
func (s *PayRatesService) GetAwardFromLibrary(ctx context.Context, awardCode string) (AwardLibraryEntry, error) {
	var award AwardLibraryEntry
	path := fmt.Sprintf("/payroll/listAwardsLibrary/%s", url.PathEscape(awardCode))
	err := s.client.do(ctx, "GET", path, nil, &award)
	return award, err
}

// OverridePayRule overrides a specific pay rule in the award library.
type OverridePayRule struct {
	Id         string  `json:"Id"`
	HourlyRate float64 `json:"HourlyRate"`
}

// SetAwardFromLibraryInput sets the award for an employee, optionally overriding pay rules.
type SetAwardFromLibraryInput struct {
	CountryCode     string            `json:"strCountryCode"`
	AwardCode       string            `json:"strAwardCode"`
	OverridePayRule []OverridePayRule `json:"arrOverridePayRules,omitempty"`
}

// SetAwardFromLibrary assigns an award from the library to an employee.
func (s *PayRatesService) SetAwardFromLibrary(ctx context.Context, employeeID int, input *SetAwardFromLibraryInput) (map[string]interface{}, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	path := fmt.Sprintf("/supervise/employee/%d/setAwardFromLibrary", employeeID)
	err = s.client.do(ctx, "POST", path, bytes.NewReader(body), &result)
	return result, err
}
