package main

import (
	"bytes"
	"errors"
	"os"
	"testing"
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

	executeFunc = func() error { return errors.New("boom") }
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
