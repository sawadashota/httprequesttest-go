// Copyright 2018 sawadashota. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package example is just sample for http client and testing.
package example

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Api struct {
	// Auth token
	token string

	// http client
	httpclient *http.Client

	// endpoint Base URL
	endpointBaseURL *url.URL
}

// New builds a API client from the provided token and options.
func New(token string, opts ...Option) *Api {
	api := &Api{
		token:      token,
		httpclient: http.DefaultClient,
		endpointBaseURL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/",
		},
	}

	for _, opt := range opts {
		opt(api)
	}

	return api
}

// Option is customize Api properties function
type Option func(*Api)

// OptionHTTPClient - provide a custom http client to the client
func OptionHTTPClient(c *http.Client) Option {
	return func(api *Api) {
		api.httpclient = c
	}
}

// EndpointURLOption .
func EndpointBaseURLOption(endpointURL *url.URL) Option {
	return func(api *Api) {
		api.endpointBaseURL = endpointURL
	}
}

// ResponseBody of example resource
type ResponseBody struct {
	Text string `json:"text"`
}

// Get example resource
func (api *Api) Get(ctx context.Context) (*ResponseBody, error) {
	req, err := http.NewRequest(http.MethodGet, api.endpointBaseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", api.token))

	resp, err := api.request(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("bad response status code %d", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var body ResponseBody
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, err
	}

	return &body, nil
}

// request .
func (api *Api) request(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)

	respCh := make(chan *http.Response)
	errCh := make(chan error)

	go func() {
		resp, err := api.httpclient.Do(req)
		if err != nil {
			errCh <- err
			return
		}

		respCh <- resp
	}()

	select {
	case resp := <-respCh:
		return resp, nil

	case err := <-errCh:
		return nil, err

	case <-ctx.Done():
		return nil, errors.New("HTTP request cancelled")
	}
}
