package main

import (
	"fmt"
	"io"
	"os"

	"github.com/salmonumbrella/deputy-cli/internal/cmd"
)

var (
	executeFunc           = cmd.Execute
	exitFunc              = os.Exit
	outWriter   io.Writer = os.Stdout
	errWriter   io.Writer = os.Stderr
)

func main() {
	result := executeFunc()
	if result.Err != nil {
		if result.JSONOutput {
			// JSON mode: structured error to stdout
			_, _ = fmt.Fprintln(outWriter, cmd.FormatErrorJSON(result.Err))
		} else {
			// Text mode: human-readable error to stderr
			_, _ = fmt.Fprintf(errWriter, "Error: %s\n", cmd.FormatError(result.Err, result.Debug))
		}
		exitFunc(result.ExitCode)
	}
}
