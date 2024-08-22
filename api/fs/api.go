package fs

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type api struct {
	client *Client
}

type payload struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

// InterceptFunc is a function that intercepts a request before it is sent.
type InterceptFunc func(context.Context, *http.Request) error

// AddInterceptor adds an interceptor to the APIClient
func (c *Client) AddInterceptor(interceptor InterceptFunc) {
	c.interceptors = append(c.interceptors, interceptor)
}

func (c *Client) prepareRequest(
	method string, path string,
	contentType string,
	body any,
) (req *http.Request, err error) {
	var (
		payload io.Reader
	)
	if body != nil {
		payload, err = parseRequestBody(body, contentType)
		if err != nil {
			return
		}
	}
	if payload != nil {
		req, err = http.NewRequest(method, path, payload)
	} else {
		req, err = http.NewRequest(method, path, nil)
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Content-Type", contentType)
	for k, v := range c.cfg.Headers {
		req.Header.Set(k, v)
	}
	if c.cfg.Host != "" {
		req.Header.Set("Host", c.cfg.Host)
	}
	return
}

// Do sends an HTTP request and returns an HTTP response.
func (c *Client) Do(ctx context.Context, req *http.Request) (res *http.Response, err error) {
	for _, interceptor := range c.interceptors {
		err = interceptor(ctx, req)
		if err != nil {
			return
		}
	}

	return c.cfg.HTTPClient.Do(req)
}

func (c *Client) decode(b []byte, v interface{}, contentType string) error {
	if strings.Contains(contentType, "application/json") {
		return json.Unmarshal(b, v)
	}
	if strings.Contains(contentType, "application/xml") {
		return xml.Unmarshal(b, v)
	}

	return errors.New("undefined response type")
}

// Set request body from an interface{}
func parseRequestBody(body interface{}, contentType string) (io.Reader, error) {
	switch data := body.(type) {
	case io.Reader:
		return data, nil
	case []byte:
		return bytes.NewBuffer(data), nil
	case string:
		return strings.NewReader(data), nil
	case *string:
		return strings.NewReader(*data), nil
	}

	var (
		bodyBuf = bytes.Buffer{}
		err     error
	)
	switch contentType {
	case "application/json":
		err = json.NewEncoder(&bodyBuf).Encode(body)
	case "application/xml":
		err = xml.NewEncoder(&bodyBuf).Encode(body)
	case "multipart/form-data":
		w := multipart.NewWriter(&bodyBuf)
		formParams, ok := body.(url.Values)
		if !ok {
			return nil, fmt.Errorf("Invalid body type %s\n", contentType)
		}
		for k, v := range formParams {
			for _, iv := range v {
				if strings.HasPrefix(k, "@") { // file
					err = addFile(w, k[1:], iv)
					if err != nil {
						return nil, err
					}
				} else { // form value
					w.WriteField(k, iv)
				}
			}
		}
		w.Close()
	default:
		err = fmt.Errorf("Invalid body type %s\n", contentType)
	}
	if err != nil {
		return nil, err
	}

	return &bodyBuf, nil
}

// Add a file to the multipart request
func addFile(w *multipart.Writer, fieldName, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	part, err := w.CreateFormFile(fieldName, filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)

	return err
}
