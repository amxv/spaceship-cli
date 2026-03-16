package main

import (
	"fmt"
	"os"

	"github.com/ashray/spaceship-cli/internal/app"
)

func main() {
	if err := app.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
