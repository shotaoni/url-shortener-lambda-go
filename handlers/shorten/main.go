package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sample/handlers/db"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
	"github.com/teris-io/shortid"
)

type request struct {
	URL string `json:"url"`
}

type Response struct {
	ShortenResource string `json:"shorten_resource`
}

type Link struct {
	ShortenResource string `json:"shorten_resource"`
	OriginalURL     string `json:"original_url"`
}

// グローバル変数でdbを宣言するとlambdaがコンテナを再利用する時に
// dbインスタンスを再利用できます
var DynamoDB db.DB

func init() {
	DynamoDB = db.New()
}

func main() {
	lambda.Start(handler)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	p, err := parseRequest(request)

	if err != nil {
		return response(
			http.StatusBadRequest,
			errorResponseBody(err.Error()),
		), nil
	}

	// 短縮URL自体はライブラリーで生成しています
	shortenResource := shortid.MustGenerate()

	// "shorten"という文字列は予約されているため、もし生成された場合は作り直します。
	for shortenResource == "shorten" {
		shortenResource = shortid.MustGenerate()
	}
	link := &Link{
		ShortenResource: shortenResource,
		OriginalURL:     p.URL,
	}

	_, err = DynamoDB.PutItem(link)
	if err != nil {
		return response(
			http.StatusInternalServerError,
			errorResponseBody(err.Error()),
		), nil
	}

	b, err := responseBody(shortenResource)
	if err != nil {
		// エラーが存在する時もreturnの第二引数でerrを返してはいけません
		// エラーを返すとどんなステータスコードも一律で502になります。
		return response(
			http.StatusInternalServerError,
			errorResponseBody(err.Error()),
		), nil
	}
	return response(http.StatusOK, b), nil
}

func parseRequest(req events.APIGatewayProxyRequest) (*request, error) {
	if req.HTTPMethod != http.MethodPost {
		return nil, fmt.Errorf("use POST request")
	}

	var r request
	err := json.Unmarshal([]byte(req.Body), &r)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse request")
	}

	// ParseRequestURIは文字列を受け取り、URLにパースするメソッド
	// エラーなくパースできることで有効なURLとみなしている
	// https://golang.org/src/net/url/url.go?s=13616:13665#L471
	_, err = url.ParseRequestURI(r.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid URL")
	}

	return &r, nil
}

func response(code int, body string) events.APIGatewayProxyResponse {
	// Lambdaプロキシ統合のレスポンスフォーマットに沿った構造体が
	// aws-lambda-goで定義されている
	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Body:       body,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

func responseBody(shortenResource string) (string, error) {
	resp, err := json.Marshal(Response{ShortenResource: shortenResource})
	if err != nil {
		return "", err
	}

	return string(resp), nil
}

func errorResponseBody(msg string) string {
	return fmt.Sprintf("{\"message\":\"%s\"}", msg)
}
