package main

import (
	"fmt"
	"net/http"
	"sample/handlers/db"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Link struct {
	ShortURL string `json:"shorten_resource"`
	LongURL  string `json:"original_url"`
}

var DynamoDB db.DB

func init() {
	DynamoDB = db.New()
}

func main() {
	lambda.Start(handler)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	r, err := parseRequest(request)
	if err != nil {
		return response(
			http.StatusBadRequest,
			errorResponseBody(err.Error()),
		), nil
	}

	URL, err := DynamoDB.GetItem(r)
	if err != nil {
		return response(
			http.StatusInternalServerError,
			errorResponseBody(err.Error()),
		), nil
	}
	if URL == "" {
		return response(
			http.StatusNotFound,
			"",
		), nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusPermanentRedirect,
		// map[key]value
		Headers: map[string]string{
			"location": URL,
		},
	}, nil
}

func parseRequest(req events.APIGatewayProxyRequest) (string, error) {
	// reqからはHTTPメソッドやクエリストロングパラメータなどが取得できる
	// 詳細は構造体定義を追ってみよう
	// https://github.com/aws/aws-lambda-go/blob/master/events/apigw.go
	if req.HTTPMethod != http.MethodGet {
		return "", fmt.Errorf("use GET request")
	}

	shortenResource := req.PathParameters["shorten_resource"]

	return shortenResource, nil
}

func response(code int, body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Body:       body,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

func errorResponseBody(msg string) string {
	return fmt.Sprintf("{\"message\":\"%s\"}", msg)
}
