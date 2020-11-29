package main

import (
	"blockchain-tutorial/cmd"
	"os"
)

func main() {
	defer os.Exit(0)
	cmd.NewCommandLine().Run()
}
