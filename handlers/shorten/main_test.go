package main

import (
	"fmt"
	"net/http"
	"os"
	"sample/handlers/db"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"
)

const exitError = 1

func TestHandler(t *testing.T) {
	tests := []struct {
		url, method string
		status      int
	}{
		{
			"https://github.com/shotaoni/url-shortener-lambda-go",
			http.MethodPost,
			http.StatusOK,
		},
		{
			"invalid URL",
			http.MethodPost,
			http.StatusBadRequest,
		},
		{
			"invalid method",
			http.MethodGet,
			http.StatusBadRequest,
		},
	}

	for _, te := range tests {
		res, _ := handler(events.APIGatewayProxyRequest{
			HTTPMethod: te.method,
			Body:       "{\"url\": \"" + te.url + "\"}",
		})

		if res.StatusCode != te.status {
			t.Errorf("ExitStatus=%d, want %d", res.StatusCode, te.status)
		}
	}
}

// 個々のテストケースの前にTestMainが実行される
// 冪等生を担保するためにhandlerごとにテーブルの作成と削除を実行する
func TestMain(m *testing.M) {
	if err := prepare(); err != nil {
		fmt.Println(err)
		os.Exit(exitError)
	}
	exitCode := m.Run()
	if err := cleanUp(); err != nil {
		fmt.Println(err)
		os.Exit(exitError)
	}
	os.Exit(exitCode)
}

func prepare() error {
	DynamoDB = db.TestNew()

	ok, err := DynamoDB.LinkTableExists()
	if err != nil {
		return errors.Wrap(err, "failed to check table existence")
	}
	if ok {
		if err := DynamoDB.DeleteLinkTable(); err != nil {
			return errors.Wrap(err, "failed to delete link table")
		}
	}
	if err := DynamoDB.CreateLinkTable(); err != nil {
		return errors.Wrap(err, "failed to create link table")
	}
	return nil
}

func cleanUp() error {
	DynamoDB = db.TestNew()

	ok, err := DynamoDB.LinkTableExists()
	if err != nil {
		return errors.Wrap(err, "failed to check table existence")
	}
	if ok {
		if err := DynamoDB.DeleteLinkTable(); err != nil {
			return errors.Wrap(err, "failed to delete link table")
		}
	}

	DynamoDB = db.DB{}

	return nil
}
