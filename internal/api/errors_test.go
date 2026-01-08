package api

import (
	"errors"
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
