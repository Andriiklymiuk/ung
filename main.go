package main

import (
	"fmt"
	"os"

	"github.com/Andriiklymiuk/ung/cmd"
	"github.com/Andriiklymiuk/ung/internal/db"
)

func main() {
	// Initialize database
	if err := db.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Execute CLI
	cmd.Execute()
}
