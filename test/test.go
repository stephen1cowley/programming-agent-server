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

	// err = awsHandlers.DeployReactApp(cfg)
	// if err != nil {
	// 	fmt.Printf("Error %v\n", err)
	// }

	err = awsHandlers.StopPreviousTask(cfg, "ProjectCluster2", "arn:aws:ecs:eu-west-2:211125355525:task/ProjectCluster2/40d7aed7d3714973893bbb6c5baa3527")
	if err != nil {
		fmt.Printf("Error stopping task %v\n", err)
	}
}
