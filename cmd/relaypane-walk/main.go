package main

import (
	"os"

	"github.com/relaypane/relaypane/internal/walkui"
)

func main() {
	if err := walkui.Run(); err != nil {
		os.Exit(1)
	}
}
