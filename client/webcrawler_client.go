package main

import (
	"io"
	"net/http"
	"net/url"
)

type wcClient struct {
	*http.Client
	baseURL string
}

func (c *wcClient) NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, c.baseURL+path, body)
}

func (c *wcClient) Get(path string) (*http.Response, error) {
	return c.Client.Get(c.baseURL + path)
}

func (c *wcClient) Head(path string) (*http.Response, error) {
	return c.Client.Head(c.baseURL + path)
}
func (c *wcClient) Post(path string, contentType string, body io.Reader) (*http.Response, error) {
	return c.Client.Post(c.baseURL+path, contentType, body)
}

func (c *wcClient) PostForm(path string, data url.Values) (*http.Response, error) {
	return c.Client.PostForm(c.baseURL+path, data)
}

func (c *wcClient) getClient(host string) (baseURL string, client *http.Client, err error) {
	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	baseURL = "http://" + host + ":8900"
	return baseURL, client, nil
}

func myClient(host string) (*wcClient, error) {
	var err error
	wcClient := wcClient{}

	wcClient.baseURL, wcClient.Client, err = wcClient.getClient(host)
	if err != nil {
		return nil, err
	}
	return &wcClient, nil
}
