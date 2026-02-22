package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type Roster struct {
	Id              int    `json:"Id"`
	Date            string `json:"Date"`
	StartTime       int64  `json:"StartTime"`
	EndTime         int64  `json:"EndTime"`
	Mealbreak       string `json:"Mealbreak"`
	Employee        int    `json:"Employee"`
	OperationalUnit int    `json:"OperationalUnit"`
	Open            bool   `json:"Open"`
	Published       bool   `json:"Published"`
	Comment         string `json:"Comment,omitempty"`
}

type RostersService struct {
	client *Client
}

func (c *Client) Rosters() *RostersService {
	return &RostersService{client: c}
}

func (s *RostersService) List(ctx context.Context, opts *ListOptions) ([]Roster, error) {
	var rosters []Roster
	err := s.client.doWithOpts(ctx, "GET", "/supervise/roster", nil, &rosters, opts)
	return rosters, err
}

func (s *RostersService) Get(ctx context.Context, id int) (*Roster, error) {
	var roster Roster
	path := fmt.Sprintf("/resource/Roster/%d", id)
	err := s.client.do(ctx, "GET", path, nil, &roster)
	return &roster, err
}

type CreateRosterInput struct {
	Employee        int    `json:"intEmployeeId"`
	OperationalUnit int    `json:"intOpunitId"`
	StartTime       int64  `json:"intStartTimestamp"`
	EndTime         int64  `json:"intEndTimestamp"`
	Mealbreak       string `json:"strMealbreak,omitempty"`
	Comment         string `json:"strComment,omitempty"`
	Open            bool   `json:"blnOpen,omitempty"`
	Publish         bool   `json:"blnPublish,omitempty"`
}

func (s *RostersService) Create(ctx context.Context, input *CreateRosterInput) (*Roster, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var roster Roster
	err = s.client.do(ctx, "POST", "/supervise/roster", bytes.NewReader(body), &roster)
	return &roster, err
}

type CopyRosterInput struct {
	FromDate string `json:"strFromDate"`
	ToDate   string `json:"strToDate"`
	Location int    `json:"intLocationId"`
}

func (s *RostersService) Copy(ctx context.Context, input *CopyRosterInput) error {
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}

	return s.client.do(ctx, "POST", "/supervise/roster/copy", bytes.NewReader(body), nil)
}

type PublishRosterInput struct {
	FromDate string `json:"strFromDate"`
	ToDate   string `json:"strToDate"`
	Location int    `json:"intLocationId"`
}

func (s *RostersService) Publish(ctx context.Context, input *PublishRosterInput) error {
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}

	return s.client.do(ctx, "POST", "/supervise/roster/publish", bytes.NewReader(body), nil)
}

func (s *RostersService) Discard(ctx context.Context, input *PublishRosterInput) error {
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}

	return s.client.do(ctx, "POST", "/supervise/roster/discard", bytes.NewReader(body), nil)
}

type SwapRoster struct {
	Id              int    `json:"Id"`
	Date            string `json:"Date"`
	StartTime       int64  `json:"StartTime"`
	EndTime         int64  `json:"EndTime"`
	Employee        int    `json:"Employee"`
	OperationalUnit int    `json:"OperationalUnit"`
}

func (s *RostersService) GetSwappable(ctx context.Context, rosterID int) ([]SwapRoster, error) {
	var rosters []SwapRoster
	path := fmt.Sprintf("/supervise/roster/%d/swap", rosterID)
	err := s.client.do(ctx, "GET", path, nil, &rosters)
	return rosters, err
}
