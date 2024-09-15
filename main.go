package main

import (
	"flag"
	"fmt"
)

func main() {
	runModeFlag := flag.String("runMode", "server", "Specify run mode (server, cli)")
	flag.Parse()
	fmt.Printf("Selected run mode: %s\n", *runModeFlag)
}
