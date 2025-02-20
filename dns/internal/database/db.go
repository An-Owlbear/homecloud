package database

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func Create(ctx context.Context) (dbClient *dynamodb.Client, err error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return
	}

	dbClient = dynamodb.NewFromConfig(cfg)
	return
}
