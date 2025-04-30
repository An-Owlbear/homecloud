package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/An-Owlbear/homecloud/dns/internal/database"
	"github.com/An-Owlbear/homecloud/dns/internal/deviceinfo"
	"github.com/An-Owlbear/homecloud/dns/internal/dns"
	"github.com/An-Owlbear/homecloud/dns/internal/lambdautils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

type Request struct {
	DeviceId  string `json:"device_id"`
	DeviceKey string `json:"device_key"`
	Port      int    `json:"port"`
}

var PortCheckError = errors.New("invalid response from port forwarded address")

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	// checks the user is a valid user and ensures
	deviceInfo, err := deviceinfo.Get(ctx, db, requestBody.DeviceId)
	if err != nil {
		return lambdautils.ErrorResponse(err, 400)
	}

	hashMatches, err := deviceinfo.CheckKey(requestBody.DeviceKey, deviceInfo.DeviceKey)
	if err != nil {
		return lambdautils.ErrorResponse(err, 500)
	}
	if !hashMatches {
		return lambdautils.ErrorResponse(deviceinfo.InvalidKeyError, 401)
	}

	hostedZone, err := dnsClient.GetHostedZone(context.Background(), &route53.GetHostedZoneInput{
		Id: aws.String(dns.HostedZoneID),
	})
	if err != nil {
		return lambdautils.ErrorResponse(err, 500)
	}
	domain, _ := strings.CutSuffix(*hostedZone.HostedZone.Name, ".")

	// Sends a HTTP request to the device's address to check port forwarding works correctly
	requestUrl := fmt.Sprintf("http://%s.%s:%d/api/v1/check", deviceInfo.Subdomain, domain, requestBody.Port)
	response, err := http.Get(requestUrl)
	if err != nil {
		return lambdautils.ErrorResponse(err, 500)
	}
	if response.StatusCode < 200 || response.StatusCode >= 500 {
		return lambdautils.ErrorResponse(PortCheckError, 500)
	}

	return lambdautils.EmptyResponse(204)
}

func main() {
	lambda.Start(handler)
}
