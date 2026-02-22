package api

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsStatus_NonAPIError(t *testing.T) {
	err := errors.New("boom")
	assert.False(t, IsStatus(err, 400))
}

func TestIsStatus_MismatchedCode(t *testing.T) {
	err := &APIError{StatusCode: 401, Message: "unauthorized"}
	assert.False(t, IsStatus(err, 403))
}

func TestIsNotFoundAndForbidden(t *testing.T) {
	assert.True(t, IsNotFound(&APIError{StatusCode: 404}))
	assert.True(t, IsForbidden(&APIError{StatusCode: 403}))
}

func TestCodeFromStatus(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{401, ErrCodeAuthRequired},
		{403, ErrCodeAuthForbidden},
		{404, ErrCodeNotFound},
		{409, ErrCodeConflict},
		{422, ErrCodeValidation},
		{429, ErrCodeRateLimited},
		{500, ErrCodeServerError},
		{502, ErrCodeServerError},
		{503, ErrCodeServerError},
		{400, ErrCodeInvalidInput},
		{418, ErrCodeInvalidInput},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.status), func(t *testing.T) {
			got := CodeFromStatus(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		status int
		want   bool
	}{
		{429, true},
		{500, true},
		{502, true},
		{503, true},
		{400, false},
		{401, false},
		{403, false},
		{404, false},
		{422, false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.status), func(t *testing.T) {
			got := IsRetryable(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}
