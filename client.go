// Package gomarathon provides a client to interact with a marathon
// api. on http or https
package gomarathon

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Client is containing the configured http.Client
// and the host url
type Client struct {
	Host       *url.URL
	HttpClient *http.Client
}

const (
	API_VERSION = "/v2"
)

// Return a pointer to the new client
func NewClient(host string, tlsConfig *tls.Config) (*Client, error) {
	// Validate url
	h, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("Can't parse host %s", host)
	}

	return &Client{
		Host:       h,
		HttpClient: newHTTPClient(h, tlsConfig),
	}, nil
}

// do the actual prepared request in request()
func (c *Client) do(method, path string, data interface{}) ([]byte, int, error) {
	var params io.Reader
	var resp *http.Response

	if data != nil {
		buf, err := json.Marshal(data)
		if err != nil {
			return nil, -1, err
		}
		params = bytes.NewBuffer(buf)
	}

	req, err := http.NewRequest(method, c.Host.String()+path, params)
	if err != nil {
		return nil, -1, err
	}

	// Prepare and do the request
	req.Header.Set("User-Agent", "gomarathon")
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.HttpClient.Do(req)
	if err != nil {
		return nil, -1, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, -1, err
	}
	if resp.StatusCode >= 400 {
		return nil, resp.StatusCode, fmt.Errorf("%d: %s", resp.StatusCode, body)
	}

	return body, resp.StatusCode, nil
}

// Prepare the request by setting the correct methods and parameters
// TODO:
// 	- find a better way to build parameters
func (c *Client) request(options *RequestOptions) (*Response, error) {

	if options.Path == "" {
		options.Path = "apps"
	}

	if options.Method == "" {
		options.Method = "GET"
	}

	path := fmt.Sprintf("%s/%s", API_VERSION, options.Path)

	if options.Params != nil {
		v := url.Values{}

		if options.Params.Cmd != "" {
			v.Set("cmd", url.QueryEscape(options.Params.Cmd))
		}

		if options.Params.Host != "" {
			v.Set("host", url.QueryEscape(options.Params.Host))
		}

		if options.Params.Scale {
			v.Set("scale", "true")
		}

		if options.Params.CallBackUrl != "" {
			v.Set("CallbackUrl", url.QueryEscape(options.Params.CallBackUrl))
		}

		path = fmt.Sprintf("%s?%s", path, v.Encode())
	}

	data, code, err := c.do(options.Method, path, options.Datas)
	if err != nil {
		return nil, err
	}
	resp := &Response{
		Code: code,
	}

	err = json.Unmarshal(data, resp)
	if err != nil {
		return resp, err
	}
	return resp, nil
}