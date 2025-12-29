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
