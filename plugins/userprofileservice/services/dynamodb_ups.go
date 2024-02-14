package services

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/optimizely/agent/plugins/userprofileservice"
	"github.com/optimizely/go-sdk/v2/pkg/decision"
	"github.com/rs/zerolog/log"
)

const (
	tableName = "UserProfileServiceTable"
)

var bgCtx = context.Background()

type DynamoDBUserProfileService struct {
	Client          *dynamodb.Client
	IAMRoleAccess   bool   `json:"iamRoleAccess"`
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
	TableName       string `json:"tableName"`
	Host            string `json:"host"`
}

func (d *DynamoDBUserProfileService) Lookup(userID string) (profile decision.UserProfile) {
	if d.Client == nil {
		d.initClient()
	}

	profile = decision.UserProfile{
		ID:                  userID,
		ExperimentBucketMap: make(map[decision.UserDecisionKey]string),
	}

	if userID == "" {
		return profile
	}

	out, err := d.Client.GetItem(bgCtx, &dynamodb.GetItemInput{
		TableName: &d.TableName,
		Key: map[string]types.AttributeValue{
			"UserID": &types.AttributeValueMemberS{Value: userID},
		},
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to get item from DynamoDB")
		return profile
	}

	if out.Item == nil {
		return profile
	}

	data, ok := out.Item["Data"].(*types.AttributeValueMemberS)
	if !ok {
		log.Error().Msg("Unexpected data type for user profile")
		return profile
	}

	experimentBucketMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(data.Value), &experimentBucketMap)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal user profile data")
	}

	prof := convertToUserProfile(map[string]interface{}{userIDKey: userID, experimentBucketMapKey: experimentBucketMap}, userIDKey)

	return prof
}

func (d *DynamoDBUserProfileService) Save(profile decision.UserProfile) {
	if d.Client == nil {
		d.initClient()
	}

	if profile.ID == "" {
		return
	}

	experimentBucketMap := map[string]interface{}{}
	for k, v := range profile.ExperimentBucketMap {
		experimentBucketMap[k.ExperimentID] = map[string]string{k.Field: v}
	}

	data, err := json.Marshal(experimentBucketMap)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal user profile data")
		return
	}

	_, err = d.Client.PutItem(bgCtx, &dynamodb.PutItemInput{
		TableName: &d.TableName,
		Item: map[string]types.AttributeValue{
			"UserID": &types.AttributeValueMemberS{Value: profile.ID},
			"Data":   &types.AttributeValueMemberS{Value: string(data)},
		},
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to put item to DynamoDB")
	}
}

func (d *DynamoDBUserProfileService) initClient() {
	var awsCfg aws.Config
	var err error

	if d.IAMRoleAccess {
		// Assume IAM role is available for authentication
		awsCfg, err = config.LoadDefaultConfig(bgCtx, config.WithRegion(d.Region))
	} else {
		awsCfg, err = config.LoadDefaultConfig(bgCtx,
			config.WithRegion(d.Region),
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{URL: d.Host}, nil
				})),
			config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID: d.AccessKeyID, SecretAccessKey: d.SecretAccessKey, SessionToken: "",
				},
			}),
		)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to load AWS SDK config")
	}

	d.Client = dynamodb.NewFromConfig(awsCfg)
}

func init() {
	log.Info().Msg("Init dynamodb ups")
	userprofileservice.Add("dynamodb", func() decision.UserProfileService {
		return &DynamoDBUserProfileService{
			TableName: tableName,
		}
	})
}
