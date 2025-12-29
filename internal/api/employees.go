package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type Employee struct {
	Id              int    `json:"Id"`
	FirstName       string `json:"FirstName"`
	LastName        string `json:"LastName"`
	DisplayName     string `json:"DisplayName"`
	Email           string `json:"Email"`
	Mobile          string `json:"Mobile"`
	Active          bool   `json:"Active"`
	Company         int    `json:"Company"`
	Role            int    `json:"Role"`
	MainAddress     int    `json:"MainAddress,omitempty"`
	Photo           any    `json:"Photo,omitempty"`
	StartDate       string `json:"StartDate,omitempty"`
	TerminationDate string `json:"TerminationDate,omitempty"`
}

type EmployeesService struct {
	client *Client
}

func (c *Client) Employees() *EmployeesService {
	return &EmployeesService{client: c}
}

func (s *EmployeesService) List(ctx context.Context, opts *ListOptions) ([]Employee, error) {
	var employees []Employee

	// Use QUERY endpoint when pagination is specified
	// Simple GET to /supervise/employee ignores query params
	if opts != nil && (opts.Limit > 0 || opts.Offset > 0) {
		input := &QueryInput{
			Max:   opts.Limit,
			Start: opts.Offset,
		}
		body, err := json.Marshal(input)
		if err != nil {
			return nil, err
		}
		err = s.client.do(ctx, "POST", "/resource/Employee/QUERY", bytes.NewReader(body), &employees)
		return employees, err
	}

	err := s.client.do(ctx, "GET", "/supervise/employee", nil, &employees)
	return employees, err
}

func (s *EmployeesService) Get(ctx context.Context, id int) (*Employee, error) {
	var employee Employee
	path := fmt.Sprintf("/supervise/employee/%d", id)
	err := s.client.do(ctx, "GET", path, nil, &employee)
	return &employee, err
}

type CreateEmployeeInput struct {
	FirstName string `json:"strFirstName"`
	LastName  string `json:"strLastName"`
	Email     string `json:"strEmail,omitempty"`
	Mobile    string `json:"strMobile,omitempty"`
	StartDate string `json:"strStartDate,omitempty"`
	Company   int    `json:"intCompany"`
	Role      int    `json:"intRoleId,omitempty"`
}

func (s *EmployeesService) Create(ctx context.Context, input *CreateEmployeeInput) (*Employee, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var employee Employee
	err = s.client.do(ctx, "POST", "/supervise/employee", bytes.NewReader(body), &employee)
	return &employee, err
}

type UpdateEmployeeInput struct {
	FirstName string `json:"strFirstName,omitempty"`
	LastName  string `json:"strLastName,omitempty"`
	Email     string `json:"strEmail,omitempty"`
	Mobile    string `json:"strMobile,omitempty"`
	Active    *bool  `json:"blnActive,omitempty"`
}

func (s *EmployeesService) Update(ctx context.Context, id int, input *UpdateEmployeeInput) (*Employee, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var employee Employee
	path := fmt.Sprintf("/resource/Employee/%d", id)
	err = s.client.do(ctx, "POST", path, bytes.NewReader(body), &employee)
	return &employee, err
}

type TerminateInput struct {
	TerminationDate string `json:"strTerminationDate"`
}

func (s *EmployeesService) Terminate(ctx context.Context, id int, date string) error {
	input := TerminateInput{TerminationDate: date}
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/supervise/employee/%d/terminate", id)
	return s.client.do(ctx, "POST", path, bytes.NewReader(body), nil)
}

func (s *EmployeesService) Invite(ctx context.Context, id int) error {
	path := fmt.Sprintf("/supervise/employee/%d/invite", id)
	return s.client.do(ctx, "POST", path, nil, nil)
}

type AssignLocationInput struct {
	Employee int `json:"intEmployeeId"`
	Location int `json:"intCompanyId"`
}

func (s *EmployeesService) AssignLocation(ctx context.Context, employeeID, locationID int) error {
	input := AssignLocationInput{
		Employee: employeeID,
		Location: locationID,
	}
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/supervise/employee/%d/location", employeeID)
	return s.client.do(ctx, "POST", path, bytes.NewReader(body), nil)
}

func (s *EmployeesService) RemoveLocation(ctx context.Context, employeeID, locationID int) error {
	path := fmt.Sprintf("/supervise/employee/%d/location/%d", employeeID, locationID)
	return s.client.do(ctx, "DELETE", path, nil, nil)
}

func (s *EmployeesService) Reactivate(ctx context.Context, id int) error {
	active := true
	input := &UpdateEmployeeInput{
		Active: &active,
	}
	_, err := s.Update(ctx, id, input)
	return err
}

func (s *EmployeesService) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/supervise/employee/%d", id)
	return s.client.do(ctx, "DELETE", path, nil, nil)
}

type Unavailability struct {
	Id        int    `json:"Id"`
	Employee  int    `json:"Employee"`
	DateStart string `json:"DateStart"`
	DateEnd   string `json:"DateEnd"`
	Comment   string `json:"Comment,omitempty"`
}

type CreateUnavailabilityInput struct {
	Employee  int    `json:"intEmployee"`
	DateStart string `json:"strDateStart"`
	DateEnd   string `json:"strDateEnd"`
	Comment   string `json:"strComment,omitempty"`
}

func (s *EmployeesService) AddUnavailability(ctx context.Context, input *CreateUnavailabilityInput) (*Unavailability, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var unavail Unavailability
	err = s.client.do(ctx, "POST", "/resource/EmployeeAvailability", bytes.NewReader(body), &unavail)
	return &unavail, err
}
