package api

import (
	"context"
	"encoding/json"
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

// MarshalJSON implements custom JSON marshaling with snake_case field names
// and an `id` field that aliases `user_id` for agent-friendliness.
func (m MeInfo) MarshalJSON() ([]byte, error) {
	type meInfoOutput struct {
		Id           int    `json:"id"`
		UserId       int    `json:"user_id"`
		EmployeeId   int    `json:"employee_id"`
		Login        string `json:"login"`
		Name         string `json:"name"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		PrimaryEmail string `json:"primary_email"`
		PrimaryPhone string `json:"primary_phone"`
		Photo        any    `json:"photo,omitempty"`
		Company      int    `json:"company"`
		Portfolio    string `json:"portfolio"`
		Role         int    `json:"role"`
	}
	return json.Marshal(meInfoOutput{
		Id:           m.UserId,
		UserId:       m.UserId,
		EmployeeId:   m.EmployeeId,
		Login:        m.Login,
		Name:         m.Name,
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		PrimaryEmail: m.PrimaryEmail,
		PrimaryPhone: m.PrimaryPhone,
		Photo:        m.Photo,
		Company:      m.Company,
		Portfolio:    m.Portfolio,
		Role:         m.Role,
	})
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
