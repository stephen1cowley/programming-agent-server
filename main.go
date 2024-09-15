package main

import (
	"flag"
	"fmt"

	cliAgent "github.com/stephen1cowley/programming-agent-server/cliAgent"
)

func main() {
	runModeFlag := flag.String("runMode", "server", "Specify run mode (server, cli)")
	flag.Parse()
	fmt.Printf("Selected run mode: %s\n", *runModeFlag)

	if *runModeFlag == "cli" {
		fmt.Println("Running CLI interface...")
		cliAgent.CliAgent()
	} else if *runModeFlag == "server" {
		fmt.Println("Setting up server...")
	}
}
