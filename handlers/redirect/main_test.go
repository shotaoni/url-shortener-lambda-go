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
		path, method string
		status       int
	}{
		{
			"xK1NKGomg",
			http.MethodGet,
			http.StatusPermanentRedirect,
		},
		{
			"xK1NKGomg",
			http.MethodPost,
			http.StatusBadRequest,
		},
		{
			"invalid path",
			http.MethodGet,
			http.StatusNotFound,
		},
	}

	for _, te := range tests {
		res, _ := handler(events.APIGatewayProxyRequest{
			PathParameters: map[string]string{"shorten_resource": te.path},
			HTTPMethod:     te.method,
		})

		if res.StatusCode != te.status {
			t.Errorf("ExitStatus=%d, want %d", res.StatusCode, te.status)
		}
	}
}

type Liink struct {
	ShortenResource string `json:"shorten_resource"`
	OriginalURL     string `json:"original_url"`
}

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

	link := &Liink{
		ShortenResource: "xK1NKGomg",
		OriginalURL:     "https://example.com/",
	}
	_, err = DynamoDB.PutItem(link)
	if err != nil {
		return errors.Wrap(err, "failed to putitem to link table")
	}
	return nil
}

func cleanUp() error {
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
