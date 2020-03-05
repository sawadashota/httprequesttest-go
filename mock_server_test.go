package example_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	example "github.com/sawadashota/httprequesttest-go"
)

// newMockServer for testing
func newMockServer() (*http.ServeMux, *url.URL) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	mockServerURL, _ := url.Parse(server.URL)
	return mux, mockServerURL
}

func TestApi_Get_MockServer(t *testing.T) {
	type response struct {
		duration time.Duration
		body     string
	}

	cases := map[string]struct {
		token          string
		requestTimeout time.Duration
		response       response
		expectedText   string
		expectHasError bool
	}{
		"valid token": {
			token:          ValidToken,
			requestTimeout: 300 * time.Millisecond,
			response: response{
				duration: 0,
				body:     "hello http test",
			},
			expectedText:   "hello http test",
			expectHasError: false,
		},
		"invalid token": {
			token:          "invalid_token",
			expectHasError: true,
		},
		"timeout": {
			token:          ValidToken,
			requestTimeout: 50 * time.Millisecond,
			response: response{
				duration: 100 * time.Millisecond,
			},
			expectHasError: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			mux, mockServerURL := newMockServer()

			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				// for timeout test
				time.Sleep(c.response.duration)

				if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", ValidToken) {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				body := example.ResponseBody{
					Text: c.response.body,
				}

				b, err := json.Marshal(body)
				if err != nil {
					t.Fatal(err)
				}

				_, _ = w.Write(b)
			})

			client := example.New(c.token, example.EndpointBaseURLOption(mockServerURL))

			ctx := context.Background()
			if c.requestTimeout > 0 {
				ctx, _ = context.WithTimeout(context.Background(), c.requestTimeout)
			}

			got, err := client.Get(ctx)
			if (err != nil) != c.expectHasError {
				t.Errorf("Get() error = %v, expectHasError %v", err, c.expectHasError)
				return
			}

			if err != nil {
				return
			}

			if got.Text != c.expectedText {
				t.Errorf("unexpected response's text. expected '%s', actual '%s'", c.expectedText, got.Text)
			}
		})
	}
}
