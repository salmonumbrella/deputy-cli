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
	if err := executeFunc(); err != nil {
		if cmd.IsJSONOutput() {
			// JSON mode: structured error to stdout
			_, _ = fmt.Fprintln(outWriter, cmd.FormatErrorJSON(err))
		} else {
			// Text mode: human-readable error to stderr
			_, _ = fmt.Fprintf(errWriter, "Error: %s\n", cmd.FormatError(err, cmd.IsDebug()))
		}
		exitFunc(1)
	}
}
