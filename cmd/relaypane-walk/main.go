package main

import (
	"fmt"
	"os"

	"github.com/relaypane/relaypane/internal/walkui"
)

// buildID is set by run-win.bat: -ldflags="-X main.buildID=<git-hash>"
var buildID = "dev"

func main() {
	fmt.Println("RelayPane Win32 walk UI — build", buildID)
	if err := walkui.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
