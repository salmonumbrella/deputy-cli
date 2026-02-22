package cmd

import (
	"errors"
	"testing"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
	"github.com/stretchr/testify/assert"
)

func TestExitCodeFromError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil error", nil, 0},
		{"generic error", errors.New("boom"), 1},
		{"auth required", &api.APIError{Code: api.ErrCodeAuthRequired, StatusCode: 401}, 3},
		{"auth forbidden", &api.APIError{Code: api.ErrCodeAuthForbidden, StatusCode: 403}, 3},
		{"not found", &api.APIError{Code: api.ErrCodeNotFound, StatusCode: 404}, 4},
		{"validation", &api.APIError{Code: api.ErrCodeValidation, StatusCode: 422}, 2},
		{"invalid input", &api.APIError{Code: api.ErrCodeInvalidInput, StatusCode: 400}, 2},
		{"invalid flag", &api.APIError{Code: api.ErrCodeInvalidFlag}, 2},
		{"rate limited", &api.APIError{Code: api.ErrCodeRateLimited, StatusCode: 429}, 5},
		{"server error", &api.APIError{Code: api.ErrCodeServerError, StatusCode: 500}, 6},
		{"timeout", &api.APIError{Code: api.ErrCodeTimeout}, 6},
		{"network error", &api.APIError{Code: api.ErrCodeNetworkError}, 6},
		{"conflict", &api.APIError{Code: api.ErrCodeConflict, StatusCode: 409}, 2},
		{"unknown flag cobra", errors.New("unknown flag: --foo"), 2},
		{"required flag cobra", errors.New("required flag(s) \"name\" not set"), 2},
		{"missing arg", errors.New("missing required argument: <id>"), 2},
		{"connection refused", errors.New("dial tcp: connection refused"), 6},
		{"no such host", errors.New("no such host"), 6},
		{"timeout plain", errors.New("context deadline exceeded (timeout)"), 6},
		{"empty code with status 401", &api.APIError{Code: "", StatusCode: 401, Message: "unauthorized"}, 3},
		{"empty result sentinel", outfmt.ErrEmptyResult, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ExitCodeFromError(tt.err))
		})
	}
}
