package lambdautils

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
)

type ErrorJson struct {
	Error string `json:"error"`
}

func JsonResponse(responseBody interface{}, status int) (events.APIGatewayProxyResponse, error) {
	bodyJson, err := json.Marshal(&responseBody)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       string(bodyJson),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func ErrorResponse(err error, status int) (events.APIGatewayProxyResponse, error) {
	return JsonResponse(ErrorJson{err.Error()}, status)
}

func EmptyResponse(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
	}, nil
}
