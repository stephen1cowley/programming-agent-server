package awsHandlers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type User struct {
	UserID   string `json:"UserID"`
	Password string `json:"Password"`
}

var dynamoClient *dynamodb.Client

// InitDynamo creates a fresh global dynamodb client
func InitDynamo(cfg aws.Config) {
	dynamoClient = dynamodb.NewFromConfig(cfg)
}

func DynamoPutUser(user User) error {
	// Marshal the user struct to a DynamoDB attribute value
	av, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	// Create the input for PutItem
	input := &dynamodb.PutItemInput{
		TableName: aws.String("programming-agent-users"),
		Item:      av,
	}

	// Put the item into the Users table
	_, err = dynamoClient.PutItem(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}
