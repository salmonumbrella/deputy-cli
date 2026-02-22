package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type Leave struct {
	Id          int     `json:"Id"`
	Employee    int     `json:"Employee"`
	Company     int     `json:"Company"`
	DateStart   string  `json:"DateStart"`
	DateEnd     string  `json:"DateEnd"`
	Status      int     `json:"Status"` // 0=awaiting, 1=approved, 2=declined, 3=cancelled, 4=pay pending, 5=pay approved
	Hours       float64 `json:"Hours"`
	Days        float64 `json:"Days"`
	ApproveBy   int     `json:"ApproveBy,omitempty"`
	PayApprover int     `json:"PayApprover,omitempty"`
	Comment     string  `json:"Comment,omitempty"`
	LeaveRule   int     `json:"LeaveRule,omitempty"`
}

type LeaveService struct {
	client *Client
}

func (c *Client) Leave() *LeaveService {
	return &LeaveService{client: c}
}

func (s *LeaveService) List(ctx context.Context, opts *ListOptions) ([]Leave, error) {
	var leaves []Leave
	err := s.client.doWithOpts(ctx, "GET", "/resource/Leave", nil, &leaves, opts)
	return leaves, err
}

func (s *LeaveService) Get(ctx context.Context, id int) (*Leave, error) {
	var leave Leave
	path := fmt.Sprintf("/resource/Leave/%d", id)
	err := s.client.do(ctx, "GET", path, nil, &leave)
	return &leave, err
}

type CreateLeaveInput struct {
	Employee  int    `json:"intEmployee"`
	DateStart string `json:"strDateStart"`
	DateEnd   string `json:"strDateEnd"`
	LeaveRule int    `json:"intLeaveRule,omitempty"`
	Comment   string `json:"strComment,omitempty"`
}

func (s *LeaveService) Create(ctx context.Context, input *CreateLeaveInput) (*Leave, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var leave Leave
	err = s.client.do(ctx, "POST", "/resource/Leave", bytes.NewReader(body), &leave)
	return &leave, err
}

type UpdateLeaveInput struct {
	Status  int    `json:"intStatus"`
	Comment string `json:"strComment,omitempty"`
}

func (s *LeaveService) Update(ctx context.Context, id int, input *UpdateLeaveInput) (*Leave, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var leave Leave
	path := fmt.Sprintf("/resource/Leave/%d", id)
	err = s.client.do(ctx, "POST", path, bytes.NewReader(body), &leave)
	return &leave, err
}

func (s *LeaveService) Approve(ctx context.Context, id int) error {
	input := UpdateLeaveInput{Status: 1}
	_, err := s.Update(ctx, id, &input)
	return err
}

func (s *LeaveService) Decline(ctx context.Context, id int, comment string) error {
	input := UpdateLeaveInput{Status: 2, Comment: comment}
	_, err := s.Update(ctx, id, &input)
	return err
}

type LeaveQueryInput struct {
	Search map[string]interface{} `json:"search,omitempty"`
	Join   []string               `json:"join,omitempty"`
	Max    int                    `json:"max,omitempty"`
	Start  int                    `json:"start,omitempty"`
}

func (s *LeaveService) Query(ctx context.Context, input *LeaveQueryInput) ([]Leave, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var leaves []Leave
	err = s.client.do(ctx, "POST", "/resource/Leave/QUERY", bytes.NewReader(body), &leaves)
	return leaves, err
}
