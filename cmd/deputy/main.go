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
	errWriter   io.Writer = os.Stderr
)

func main() {
	if err := executeFunc(); err != nil {
		_, _ = fmt.Fprintf(errWriter, "Error: %s\n", cmd.FormatError(err, cmd.IsDebug()))
		exitFunc(1)
	}
}
