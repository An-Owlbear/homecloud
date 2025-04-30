package lambdautils

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

type ErrorJson struct {
	Error string `json:"error"`
}

// JsonResponse encodes the given information as JSON, and sends a response with the correct headers
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

// ErrorResponse sends an error JSON response for the given error
func ErrorResponse(err error, status int) (events.APIGatewayProxyResponse, error) {
	return JsonResponse(ErrorJson{err.Error()}, status)
}

// EmptyResponse sends a response, that after going through the AWS API gateway and reaching the client, will be empty
func EmptyResponse(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
	}, nil
}
