package main

import (
	"fmt"

	apiAgent "github.com/stephen1cowley/programming-agent-server/apiAgent"
)

func main() {
	fmt.Println("Running API server...")
	apiAgent.ApiAgent()
}
