package awsHandlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

func getECRLogin(cfg aws.Config) error {
	client := ecr.NewFromConfig(cfg)
	result, err := client.GetAuthorizationToken(context.TODO(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return err
	}

	authData := result.AuthorizationData[0]
	decodedToken, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
	if err != nil {
		return err
	}

	credentials := strings.Split(string(decodedToken), ":")
	registry := *authData.ProxyEndpoint

	cmd := exec.Command("docker", "login", "--username", credentials[0], "--password", credentials[1], registry)
	return cmd.Run()
}

func buildDockerImage(imageName string) error {
	err := os.Chdir("/home/ubuntu/my-react-app")
	if err != nil {
		fmt.Println("Error changing directory:", err)
		return err
	}
	cmd := exec.Command("sudo", "docker", "build", "-t", imageName, ".")

	// Capture the combined stdout and stderr output
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing docker build: %v\n", err)
		fmt.Printf("Docker build output: %s\n", string(output)) // Output error details
		return err
	}

	fmt.Printf("Docker build output: %s\n", string(output))
	return nil
}

func pushDockerImage(imageName, ecrRepo string) error {
	// cdCmd := exec.Command("cd", "~/my-react-app")
	// if err := cdCmd.Run(); err != nil {
	// 	return err
	// }
	tagCmd := exec.Command("docker", "tag", imageName, ecrRepo)
	if err := tagCmd.Run(); err != nil {
		return err
	}
	pushCmd := exec.Command("docker", "push", ecrRepo)
	return pushCmd.Run()
}
