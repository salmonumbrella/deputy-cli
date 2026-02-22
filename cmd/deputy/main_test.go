package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/cmd"
)

func TestMain_ExecutesVersion(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"deputy", "version"}
	main()
}

func TestMain_ErrorPath(t *testing.T) {
	oldExecute := executeFunc
	oldExit := exitFunc
	oldErrWriter := errWriter
	defer func() {
		executeFunc = oldExecute
		exitFunc = oldExit
		errWriter = oldErrWriter
	}()

	executeFunc = func() cmd.ExecuteResult {
		err := errors.New("boom")
		return cmd.ExecuteResult{Err: err, ExitCode: cmd.ExitCodeFromError(err)}
	}
	var exitCode int
	exitFunc = func(code int) { exitCode = code }

	buf := &bytes.Buffer{}
	errWriter = buf

	main()

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if buf.String() == "" {
		t.Fatal("expected error output")
	}
}

func TestMain_JSONErrorOutput(t *testing.T) {
	// Save originals
	origExecute := executeFunc
	origExit := exitFunc
	origOut := outWriter
	origErr := errWriter
	defer func() {
		executeFunc = origExecute
		exitFunc = origExit
		outWriter = origOut
		errWriter = origErr
	}()

	// Capture output
	var stdout, stderr bytes.Buffer
	outWriter = &stdout
	errWriter = &stderr

	var exitCode int
	exitFunc = func(code int) { exitCode = code }

	// Simulate JSON-mode API error via ExecuteResult
	apiErr := &api.APIError{
		Code:       api.ErrCodeNotFound,
		StatusCode: 404,
		Message:    "employee not found",
	}
	executeFunc = func() cmd.ExecuteResult {
		return cmd.ExecuteResult{
			Err:        apiErr,
			ExitCode:   cmd.ExitCodeFromError(apiErr),
			JSONOutput: true,
		}
	}

	main()

	if exitCode != 4 {
		t.Errorf("exit code = %d, want 4", exitCode)
	}

	// Stderr should be empty in JSON mode
	if stderr.Len() > 0 {
		t.Errorf("stderr should be empty in JSON mode, got: %s", stderr.String())
	}

	// Stdout should have JSON error
	var jsonErr cmd.JSONError
	if err := json.Unmarshal(stdout.Bytes(), &jsonErr); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nGot: %s", err, stdout.String())
	}

	if jsonErr.Error.Code != "NOT_FOUND" {
		t.Errorf("error code = %q, want NOT_FOUND", jsonErr.Error.Code)
	}
	if jsonErr.Error.Status != 404 {
		t.Errorf("error status = %d, want 404", jsonErr.Error.Status)
	}
}

func TestMain_ExitCodeMapping(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		json     bool
		wantCode int
	}{
		{"auth error", &api.APIError{Code: api.ErrCodeAuthRequired, StatusCode: 401, Message: "unauthorized"}, false, 3},
		{"not found json", &api.APIError{Code: api.ErrCodeNotFound, StatusCode: 404, Message: "not found"}, true, 4},
		{"rate limited", &api.APIError{Code: api.ErrCodeRateLimited, StatusCode: 429, Message: "slow down"}, false, 5},
		{"server error", &api.APIError{Code: api.ErrCodeServerError, StatusCode: 500, Message: "internal"}, false, 6},
		{"generic error", errors.New("boom"), false, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origExecute := executeFunc
			origExit := exitFunc
			origOut := outWriter
			origErr := errWriter
			defer func() {
				executeFunc = origExecute
				exitFunc = origExit
				outWriter = origOut
				errWriter = origErr
			}()

			outWriter = &bytes.Buffer{}
			errWriter = &bytes.Buffer{}

			var exitCode int
			exitFunc = func(code int) { exitCode = code }
			executeFunc = func() cmd.ExecuteResult {
				return cmd.ExecuteResult{
					Err:        tt.err,
					ExitCode:   cmd.ExitCodeFromError(tt.err),
					JSONOutput: tt.json,
				}
			}

			main()

			if exitCode != tt.wantCode {
				t.Errorf("exit code = %d, want %d", exitCode, tt.wantCode)
			}
		})
	}
}
