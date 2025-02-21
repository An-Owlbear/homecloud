package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/An-Owlbear/homecloud/dns/internal/database"
	"github.com/An-Owlbear/homecloud/dns/internal/deviceinfo"
	"github.com/An-Owlbear/homecloud/dns/internal/dns"
	"github.com/An-Owlbear/homecloud/dns/internal/lambdautils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var InvalidKeyError = errors.New("invalid key")
var SubdomainTakenError = errors.New("subdomain already in use")

type Request struct {
	DeviceId  string `json:"device_id"`
	DeviceKey string `json:"device_key"`
	Subdomain string `json:"subdomain"`
	IPAddress string `json:"ip_address"`
}

// Updates the subdomain for the given device id and key. A subdomain that is assigned to another
// domain can't be used. Uses API Gateway request/response structs to work properly with function URLs
// or AWS API gateway
func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Creates dynamodb and route53 client
	db, err := database.Create(ctx)
	if err != nil {
		return lambdautils.ErrorResponse(err, 500)
	}

	dnsClient, err := dns.New(ctx)
	if err != nil {
		return lambdautils.ErrorResponse(err, 500)
	}

	var requestBody Request
	if err := json.Unmarshal([]byte(request.Body), &requestBody); err != nil {
		return lambdautils.ErrorResponse(err, 400)
	}

	// Retrieves device info, ensures key matches
	deviceInfo, err := deviceinfo.Get(ctx, db, requestBody.DeviceId)
	if err != nil {
		return lambdautils.ErrorResponse(err, 400)
	}

	hashMatches, err := deviceinfo.CheckKey(requestBody.DeviceKey, deviceInfo.DeviceKey)
	if err != nil {
		return lambdautils.ErrorResponse(err, 500)
	}

	if !hashMatches {
		return lambdautils.ErrorResponse(InvalidKeyError, 401)
	}

	// Sets subdomain if another device isn't already using it
	subdomainTaken, err := deviceinfo.SubdomainTaken(ctx, db, requestBody.DeviceId, requestBody.Subdomain)
	if err != nil {
		return lambdautils.ErrorResponse(err, 500)
	}

	if subdomainTaken {
		return lambdautils.ErrorResponse(SubdomainTakenError, 409)
	}

	if deviceInfo.Subdomain != "" {
		err = dns.RemoveRecord(ctx, dnsClient, deviceInfo.Subdomain, requestBody.IPAddress)
		if err != nil {
			return lambdautils.ErrorResponse(err, 500)
		}
	}

	deviceInfo.Subdomain = requestBody.Subdomain
	err = deviceinfo.Put(ctx, db, deviceInfo)
	if err != nil {
		return lambdautils.ErrorResponse(err, 500)
	}

	err = dns.SetRecord(ctx, dnsClient, deviceInfo.Subdomain, requestBody.IPAddress)
	if err != nil {
		return lambdautils.ErrorResponse(err, 500)
	}

	return lambdautils.EmptyResponse(204)
}

func main() {
	lambda.Start(handler)
}
