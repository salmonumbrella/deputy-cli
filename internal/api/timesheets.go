package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type Timesheet struct {
	Id              int     `json:"Id"`
	Employee        int     `json:"Employee"`
	Date            string  `json:"Date"`
	StartTime       int64   `json:"StartTime"`
	EndTime         int64   `json:"EndTime"`
	Mealbreak       string  `json:"Mealbreak"`
	TotalTime       float64 `json:"TotalTime"`
	TotalTimeStr    string  `json:"TotalTimeStr"`
	OperationalUnit int     `json:"OperationalUnit"`
	IsInProgress    bool    `json:"IsInProgress"`
	IsLeave         bool    `json:"IsLeave"`
	Comment         string  `json:"Comment,omitempty"`
	Cost            float64 `json:"Cost"`
}

type TimesheetsService struct {
	client *Client
}

func (c *Client) Timesheets() *TimesheetsService {
	return &TimesheetsService{client: c}
}

func (s *TimesheetsService) List(ctx context.Context, opts *ListOptions) ([]Timesheet, error) {
	var timesheets []Timesheet
	err := s.client.doWithOpts(ctx, "GET", "/my/timesheets", nil, &timesheets, opts)
	return timesheets, err
}

func (s *TimesheetsService) Query(ctx context.Context, input *QueryInput) ([]Timesheet, error) {
	results, err := s.client.Resource("Timesheet").Query(ctx, input)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}

	var timesheets []Timesheet
	if err := json.Unmarshal(payload, &timesheets); err != nil {
		return nil, err
	}

	return timesheets, nil
}

func (s *TimesheetsService) Get(ctx context.Context, id int) (*Timesheet, error) {
	var timesheet Timesheet
	path := fmt.Sprintf("/supervise/timesheet/%d", id)
	err := s.client.do(ctx, "GET", path, nil, &timesheet)
	return &timesheet, err
}

type ClockInput struct {
	Employee        int    `json:"intEmployeeId,omitempty"`
	Timesheet       int    `json:"intTimesheetId,omitempty"`
	OperationalUnit int    `json:"intOpunitId,omitempty"`
	Comment         string `json:"strComment,omitempty"`
}

type ClockResponse struct {
	Id       int `json:"Id"`
	Employee int `json:"Employee"`
}

func (s *TimesheetsService) ClockIn(ctx context.Context, input *ClockInput) (*ClockResponse, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var resp ClockResponse
	err = s.client.do(ctx, "POST", "/supervise/timesheet/start", bytes.NewReader(body), &resp)
	return &resp, err
}

func (s *TimesheetsService) ClockOut(ctx context.Context, input *ClockInput) (*ClockResponse, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var resp ClockResponse
	err = s.client.do(ctx, "POST", "/supervise/timesheet/stop", bytes.NewReader(body), &resp)
	return &resp, err
}

func (s *TimesheetsService) StartBreak(ctx context.Context, input *ClockInput) error {
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}

	return s.client.do(ctx, "POST", "/supervise/timesheet/pause", bytes.NewReader(body), nil)
}

func (s *TimesheetsService) EndBreak(ctx context.Context, input *ClockInput) error {
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}

	return s.client.do(ctx, "POST", "/supervise/timesheet/resume", bytes.NewReader(body), nil)
}

type UpdateTimesheetInput struct {
	Cost *float64 `json:"Cost,omitempty"`
}

func (s *TimesheetsService) Update(ctx context.Context, id int, input *UpdateTimesheetInput) (*Timesheet, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var timesheet Timesheet
	path := fmt.Sprintf("/resource/Timesheet/%d", id)
	err = s.client.do(ctx, "POST", path, bytes.NewReader(body), &timesheet)
	return &timesheet, err
}

// PayRule represents a pay rule in Deputy
type PayRule struct {
	Id         int     `json:"Id"`
	PayTitle   string  `json:"PayTitle"`
	HourlyRate float64 `json:"HourlyRate"`
}

// TimesheetPayReturn links a pay rule to a timesheet
type TimesheetPayReturn struct {
	Id         int     `json:"Id"`
	Timesheet  int     `json:"Timesheet"`
	PayRule    int     `json:"PayRule"`
	Value      float64 `json:"Value"`
	Cost       float64 `json:"Cost"`
	Overridden bool    `json:"Overridden"`
}

// ListPayRules returns all pay rules, optionally filtered by hourly rate
func (s *TimesheetsService) ListPayRules(ctx context.Context, hourlyRate *float64) ([]PayRule, error) {
	input := &QueryInput{}
	if hourlyRate != nil {
		input.Search = map[string]interface{}{
			"s1": map[string]interface{}{"field": "HourlyRate", "type": "eq", "data": *hourlyRate},
		}
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var rules []PayRule
	err = s.client.do(ctx, "POST", "/resource/PayRules/QUERY", bytes.NewReader(body), &rules)
	return rules, err
}

// GetPayReturn gets the pay return record for a timesheet
func (s *TimesheetsService) GetPayReturn(ctx context.Context, timesheetID int) (*TimesheetPayReturn, error) {
	input := &QueryInput{
		Search: map[string]interface{}{
			"s1": map[string]interface{}{"field": "Timesheet", "type": "eq", "data": timesheetID},
		},
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var returns []TimesheetPayReturn
	err = s.client.do(ctx, "POST", "/resource/TimesheetPayReturn/QUERY", bytes.NewReader(body), &returns)
	if err != nil {
		return nil, err
	}
	if len(returns) == 0 {
		return nil, fmt.Errorf("no pay return found for timesheet %d", timesheetID)
	}
	return &returns[0], nil
}

// SetPayRuleInput for updating a timesheet's pay rule
type SetPayRuleInput struct {
	PayRule    int     `json:"PayRule"`
	Cost       float64 `json:"Cost"`
	Overridden bool    `json:"Overridden"`
}

// SetPayRule sets the pay rule for a timesheet
func (s *TimesheetsService) SetPayRule(ctx context.Context, timesheetID int, payRuleID int) (*TimesheetPayReturn, error) {
	// First get the existing pay return to get its ID
	existing, err := s.GetPayReturn(ctx, timesheetID)
	if err != nil {
		return nil, err
	}

	// Get the timesheet to calculate cost
	timesheet, err := s.Get(ctx, timesheetID)
	if err != nil {
		return nil, err
	}

	// Get the pay rule to get its hourly rate
	var rules []PayRule
	ruleInput := &QueryInput{
		Search: map[string]interface{}{
			"s1": map[string]interface{}{"field": "Id", "type": "eq", "data": payRuleID},
		},
	}
	ruleBody, err := json.Marshal(ruleInput)
	if err != nil {
		return nil, err
	}
	if err := s.client.do(ctx, "POST", "/resource/PayRules/QUERY", bytes.NewReader(ruleBody), &rules); err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("pay rule %d not found", payRuleID)
	}

	// Validate timesheet has hours
	if timesheet.TotalTime <= 0 {
		return nil, fmt.Errorf("timesheet %d has no hours recorded (TotalTime: %.2f)", timesheetID, timesheet.TotalTime)
	}

	// Calculate cost: hourly rate × hours
	cost := rules[0].HourlyRate * timesheet.TotalTime

	// Update the pay return (this sets the pay rule selection)
	input := &SetPayRuleInput{
		PayRule:    payRuleID,
		Cost:       cost,
		Overridden: true,
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var result TimesheetPayReturn
	path := fmt.Sprintf("/resource/TimesheetPayReturn/%d", existing.Id)
	if err := s.client.do(ctx, "POST", path, bytes.NewReader(body), &result); err != nil {
		return nil, err
	}

	// Also update the timesheet's cost directly (required for cost to sync)
	tsInput := &UpdateTimesheetInput{Cost: &cost}
	tsBody, err := json.Marshal(tsInput)
	if err != nil {
		return nil, err
	}
	tsPath := fmt.Sprintf("/resource/Timesheet/%d", timesheetID)
	if err := s.client.do(ctx, "POST", tsPath, bytes.NewReader(tsBody), nil); err != nil {
		return nil, err
	}

	return &result, nil
}
