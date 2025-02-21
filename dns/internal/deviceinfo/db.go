package deviceinfo

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"os"
)

var TableName = os.Getenv("TABLE_NAME")

var NotFoundError = errors.New("key not found")

func Get(ctx context.Context, client *dynamodb.Client, deviceId string) (deviceInfo DeviceInfo, err error) {
	idAttributeValue, err := attributevalue.Marshal(deviceId)
	if err != nil {
		return
	}

	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key: map[string]types.AttributeValue{
			"device_id": idAttributeValue,
		},
	})

	if err != nil {
		return
	}

	if result.Item == nil {
		return
	}

	err = attributevalue.UnmarshalMap(result.Item, &deviceInfo)
	return
}

func SubdomainTaken(ctx context.Context, client *dynamodb.Client, deviceId string, subdomain string) (bool, error) {
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		IndexName:              aws.String("subdomain-index"),
		KeyConditionExpression: aws.String("subdomain = :subdomain"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":subdomain": &types.AttributeValueMemberS{Value: subdomain},
		},
		Limit: aws.Int32(1),
	})

	if err != nil {
		return false, err
	}

	if len(result.Items) > 0 {
		var deviceInfo SubdomainIndex
		err = attributevalue.UnmarshalMap(result.Items[0], &deviceInfo)
		if err != nil {
			return false, err
		}

		if deviceInfo.DeviceId != deviceId {
			return true, nil
		}
	}

	return false, nil
}

func Put(ctx context.Context, client *dynamodb.Client, deviceInfo DeviceInfo) error {
	deviceInfoMap, err := attributevalue.MarshalMap(deviceInfo)
	if err != nil {
		return err
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(TableName),
		Item:      deviceInfoMap,
	})
	if err != nil {
		return err
	}

	return nil
}
