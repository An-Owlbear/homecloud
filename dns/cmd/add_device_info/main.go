package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"github.com/An-Owlbear/homecloud/dns/internal/database"
	"github.com/An-Owlbear/homecloud/dns/internal/deviceinfo"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
)

type Response struct {
	DeviceId  string `json:"device_id"`
	DeviceKey string `json:"device_key"`
}

// Creates a new device ID and key for a subdomain to be assigned to
// The key is returned unhashed, this must be stored in the device image
func handler(ctx context.Context) (response *Response, err error) {
	db, err := database.Create(ctx)
	if err != nil {
		return nil, err
	}

	deviceId := uuid.New().String()

	// creates random device key and hashes it
	deviceKeyBytes := make([]byte, 16)
	_, err = rand.Read(deviceKeyBytes)
	if err != nil {
		return nil, err
	}
	deviceKey := base64.StdEncoding.EncodeToString(deviceKeyBytes)

	hashedKey, err := deviceinfo.HashKey(deviceKey)
	if err != nil {
		return nil, err
	}

	err = deviceinfo.Put(ctx, db, deviceinfo.DeviceInfo{
		DeviceId:  deviceId,
		DeviceKey: hashedKey,
	})
	if err != nil {
		return nil, err
	}

	return &Response{DeviceId: deviceId, DeviceKey: deviceKey}, nil
}

func main() {
	lambda.Start(handler)
}
