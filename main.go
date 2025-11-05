package main

import (
	"os"
)

func main() {
	root := setupCommands()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
