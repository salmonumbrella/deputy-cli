package api

import (
	"context"
)

type MeInfo struct {
	UserId       int    `json:"UserId"`
	EmployeeId   int    `json:"EmployeeId"`
	Login        string `json:"Login"`
	Name         string `json:"Name"`
	FirstName    string `json:"FirstName"`
	LastName     string `json:"LastName"`
	PrimaryEmail string `json:"PrimaryEmail"`
	PrimaryPhone string `json:"PrimaryPhone"`
	Photo        any    `json:"Photo,omitempty"`
	Company      int    `json:"Company"`
	Portfolio    string `json:"Portfolio"`
	Role         int    `json:"Role"`
}

type MeService struct {
	client *Client
}

func (c *Client) Me() *MeService {
	return &MeService{client: c}
}

func (s *MeService) Info(ctx context.Context) (*MeInfo, error) {
	var info MeInfo
	err := s.client.do(ctx, "GET", "/me", nil, &info)
	return &info, err
}

func (s *MeService) Timesheets(ctx context.Context) ([]Timesheet, error) {
	var timesheets []Timesheet
	err := s.client.do(ctx, "GET", "/my/timesheets", nil, &timesheets)
	return timesheets, err
}

func (s *MeService) Rosters(ctx context.Context) ([]Roster, error) {
	var rosters []Roster
	err := s.client.do(ctx, "GET", "/my/rosters", nil, &rosters)
	return rosters, err
}

func (s *MeService) Leave(ctx context.Context) ([]Leave, error) {
	var leaves []Leave
	err := s.client.do(ctx, "GET", "/my/leave", nil, &leaves)
	return leaves, err
}
