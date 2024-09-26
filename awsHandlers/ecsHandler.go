package awsHandlers

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// registerTaskDefinition registers an ECS task definition for Fargate using the provided ECR image.
func registerTaskDefinition(cfg aws.Config, taskDefinitionName, ecrImage string) (*ecs.RegisterTaskDefinitionOutput, error) {
	client := ecs.NewFromConfig(cfg)

	input := &ecs.RegisterTaskDefinitionInput{
		Family: aws.String(taskDefinitionName),
		ContainerDefinitions: []ecstypes.ContainerDefinition{
			{
				Name:      aws.String(taskDefinitionName),
				Image:     aws.String(ecrImage),
				Essential: aws.Bool(true),
				Memory:    aws.Int32(512),
				Cpu:       256,
				PortMappings: []ecstypes.PortMapping{
					{
						ContainerPort: aws.Int32(80),
						HostPort:      aws.Int32(80),
					},
				},
			},
		},
		NetworkMode:             ecstypes.NetworkModeAwsvpc,
		RequiresCompatibilities: []ecstypes.Compatibility{ecstypes.CompatibilityFargate},
		Cpu:                     aws.String("256"),
		Memory:                  aws.String("512"),
	}

	return client.RegisterTaskDefinition(context.TODO(), input)
}

func runFargateTask(cfg aws.Config, clusterName, taskDefinitionName, subnetID, securityGroupID string) (*ecs.RunTaskOutput, error) {
	client := ecs.NewFromConfig(cfg)

	input := &ecs.RunTaskInput{
		Cluster:        aws.String(clusterName),
		TaskDefinition: aws.String(taskDefinitionName),
		LaunchType:     ecstypes.LaunchTypeFargate,
		NetworkConfiguration: &ecstypes.NetworkConfiguration{
			AwsvpcConfiguration: &ecstypes.AwsVpcConfiguration{
				Subnets:        []string{subnetID},
				SecurityGroups: []string{securityGroupID},
				AssignPublicIp: ecstypes.AssignPublicIpEnabled,
			},
		},
	}

	return client.RunTask(context.TODO(), input)
}

func deployReactApp(cfg aws.Config) error {
	imageName := "my-test-image"
	ecrRepo := "211125355525.dkr.ecr.eu-west-2.amazonaws.com/programming-agent-ui:latest"
	clusterName := ""
	subnetID := ""
	securityGroupID := ""

	err := getECRLogin(cfg)
	if err != nil {
		return err
	}

	// Build, push, and deploy
	if err := buildDockerImage(imageName); err != nil {
		return err
	}

	if err := pushDockerImage(imageName, ecrRepo); err != nil {
		return err
	}

	_, err = registerTaskDefinition(cfg, imageName, ecrRepo)
	if err != nil {
		return err
	}

	_, err = runFargateTask(cfg, clusterName, imageName, subnetID, securityGroupID)
	if err != nil {
		return err
	}

	return nil
}
