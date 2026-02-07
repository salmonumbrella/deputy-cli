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
		return cmd.ExecuteResult{Err: errors.New("boom")}
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
	executeFunc = func() cmd.ExecuteResult {
		return cmd.ExecuteResult{
			Err: &api.APIError{
				Code:       api.ErrCodeNotFound,
				StatusCode: 404,
				Message:    "employee not found",
			},
			JSONOutput: true,
		}
	}

	main()

	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1", exitCode)
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
