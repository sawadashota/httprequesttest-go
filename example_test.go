package example_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	example "github.com/sawadashota/httprequesttest-go"
)

// Auth ValidToken
const ValidToken = "valid_token"

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

// client for testing normal response
// *http.Response: only give value when you expectedText customize reponse
func client(t *testing.T, respTime time.Duration, resp *http.Response) *http.Client {
	t.Helper()

	body := example.ResponseBody{
		Text: "hello",
	}

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	return NewTestClient(func(req *http.Request) *http.Response {
		time.Sleep(respTime)

		if resp != nil {
			return resp
		}

		if req.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", ValidToken) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       nil,
				Header:     make(http.Header),
			}
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(b)),
			Header:     make(http.Header),
		}
	})
}

func TestApi_Get_MockResponse(t *testing.T) {
	cases := map[string]struct {
		token                string
		client               *http.Client
		expectHasError       bool
		expectedErrorMessage string
		expectedText         string
	}{
		"normal": {
			token:          ValidToken,
			client:         client(t, 0, nil),
			expectHasError: false,
			expectedText:   "hello",
		},
		"invalid token": {
			token:                "invalid_token",
			client:               client(t, 0, nil),
			expectHasError:       true,
			expectedErrorMessage: "bad response status code 401",
		},
		"internal server error response": {
			token: ValidToken,
			client: client(t, 0, &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       nil,
				Header:     make(http.Header),
			}),
			expectHasError:       true,
			expectedErrorMessage: "bad response status code 500",
		},
		"plain text response": {
			token: ValidToken,
			client: client(t, 0, &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString("bad")),
				Header:     make(http.Header),
			}),
			expectHasError:       true,
			expectedErrorMessage: "invalid character 'b' looking for beginning of value",
		},
		"long response time": {
			token:                ValidToken,
			client:               client(t, 3*time.Second, nil),
			expectHasError:       true,
			expectedErrorMessage: "HTTP request cancelled",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			e := example.New(c.token, example.OptionHTTPClient(c.client))

			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			resp, err := e.Get(ctx)

			if c.expectHasError {
				if err == nil {
					t.Errorf("expected error but no errors ouccured")
					return
				}

				if err.Error() != c.expectedErrorMessage {
					t.Errorf("unexpected error message. expected '%s', actual '%s'", c.expectedErrorMessage, err.Error())
				}

				return
			}

			if err != nil {
				t.Errorf(err.Error())
				return // because when has error, response is nil
			}

			if resp.Text != c.expectedText {
				t.Errorf("unexpected response's text. expected '%s', actual '%s'", c.expectedText, resp.Text)
			}
		})
	}
}
