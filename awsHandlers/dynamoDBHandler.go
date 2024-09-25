package awsHandlers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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

// DynamoPutUser creates a new user with the given details
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

// DynamoGetUser retrieves a user's information from the DynamoDB table based on the UserID
func DynamoGetUser(userID string) (*User, error) {
	// Create the input for GetItem
	input := &dynamodb.GetItemInput{
		TableName: aws.String("programming-agent-users"),
		Key: map[string]types.AttributeValue{
			"UserID": &types.AttributeValueMemberS{Value: userID}, // Assuming UserID is the primary key
		},
	}

	// Get the item from the table
	result, err := dynamoClient.GetItem(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	// Check if the item exists
	if result.Item == nil {
		return nil, fmt.Errorf("user with ID %s not found", userID)
	}

	// Unmarshal the result into a User struct
	var user User
	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}
