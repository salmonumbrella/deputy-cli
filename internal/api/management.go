package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type Memo struct {
	Id        int    `json:"Id"`
	Content   string `json:"Content"`
	Company   int    `json:"Company"`
	Creator   int    `json:"Creator"`
	Created   int64  `json:"Created"`
	ShowFrom  int64  `json:"ShowFrom,omitempty"`
	ShowUntil int64  `json:"ShowUntil,omitempty"`
}

type Journal struct {
	Id       int    `json:"Id"`
	Employee int    `json:"Employee"`
	Company  int    `json:"Company"`
	Comment  string `json:"Comment"`
	Created  int64  `json:"Created"`
	Category int    `json:"Category,omitempty"`
}

type ManagementService struct {
	client *Client
}

func (c *Client) Management() *ManagementService {
	return &ManagementService{client: c}
}

type CreateMemoInput struct {
	Content   string `json:"strContent"`
	Company   int    `json:"intCompanyId"`
	ShowFrom  int64  `json:"intShowFrom,omitempty"`
	ShowUntil int64  `json:"intShowUntil,omitempty"`
	Locations []int  `json:"arrLocation,omitempty"`
	Employees []int  `json:"arrEmployee,omitempty"`
}

func (s *ManagementService) CreateMemo(ctx context.Context, input *CreateMemoInput) (*Memo, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var memo Memo
	err = s.client.do(ctx, "PUT", "/supervise/memo", bytes.NewReader(body), &memo)
	return &memo, err
}

func (s *ManagementService) ListMemos(ctx context.Context, companyID int) ([]Memo, error) {
	var memos []Memo
	path := fmt.Sprintf("/supervise/memo?company=%d", companyID)
	err := s.client.do(ctx, "GET", path, nil, &memos)
	return memos, err
}

type CreateJournalInput struct {
	Employee int    `json:"intEmployeeId"`
	Company  int    `json:"intCompanyId"`
	Comment  string `json:"strComment"`
	Category int    `json:"intCategory,omitempty"`
}

func (s *ManagementService) PostJournal(ctx context.Context, input *CreateJournalInput) (*Journal, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var journal Journal
	err = s.client.do(ctx, "POST", "/supervise/journal", bytes.NewReader(body), &journal)
	return &journal, err
}

func (s *ManagementService) ListJournals(ctx context.Context, employeeID int) ([]Journal, error) {
	var journals []Journal
	path := fmt.Sprintf("/supervise/journal?employee=%d", employeeID)
	err := s.client.do(ctx, "GET", path, nil, &journals)
	return journals, err
}
