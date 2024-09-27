package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	awsHandlers "github.com/stephen1cowley/programming-agent-server/awsHandlers"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-west-2"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	err = awsHandlers.DeployReactApp(cfg)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
}
