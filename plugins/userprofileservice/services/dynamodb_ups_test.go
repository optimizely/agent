package services

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/optimizely/go-sdk/v2/pkg/decision"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testTableName       = "UserProfileServiceTable" // Ensure this matches your DynamoDB table name
	testRegion          = "us-west-2"               // Update this to your DynamoDB table's region
	testAccessKeyID     = "test"                    // For local testing, can be dummy
	testSecretAccessKey = "test"                    // For local testing, can be dummy
)

func TestDynamoDBUserProfileServiceIntegration(t *testing.T) {
	ctx := context.TODO()

	dynamoDBHost := "http://localhost:8000"

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(testRegion),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: dynamoDBHost}, nil
			})),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(testAccessKeyID, testSecretAccessKey, "")),
	)
	assert.NoError(t, err)

	dynamoDBClient := dynamodb.NewFromConfig(awsCfg)
	ups := DynamoDBUserProfileService{
		Client:          dynamoDBClient,
		TableName:       testTableName,
		IAMRoleAccess:   false, // Assuming static credentials for integration tests
		AccessKeyID:     testAccessKeyID,
		SecretAccessKey: testSecretAccessKey,
		Region:          testRegion,
		Host:            dynamoDBHost,
	}

	userID := "userIDValue"
	decisionKey := decision.NewUserDecisionKey("experimentId")
	decisionValue := "experimentIdValue"

	// Test Save
	profile := decision.UserProfile{
		ID: userID,
		ExperimentBucketMap: map[decision.UserDecisionKey]string{
			decisionKey: decisionValue,
		},
	}
	ups.Save(profile)

	// Test Lookup
	retrievedProfile := ups.Lookup(userID)
	assert.Equal(t, userID, retrievedProfile.ID)
	assert.Equal(t, decisionValue, retrievedProfile.ExperimentBucketMap[decisionKey])

	//Clean up test data
	_, err = dynamoDBClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(testTableName),
		Key: map[string]types.AttributeValue{
			"UserID": &types.AttributeValueMemberS{Value: userID},
		},
	})
	assert.NoError(t, err)
}
