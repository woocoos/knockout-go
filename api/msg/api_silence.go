// Code generated by woco, DO NOT EDIT.

package msg

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type SilenceAPI api

// (DELETE /silence/{silenceID})
func (a *SilenceAPI) DeleteSilence(ctx context.Context, req *DeleteSilenceRequest) (resp *http.Response, err error) {
	var (
		contentType string
		body        any
	)
	path := "/silence/{silenceID}"
	path = path[:9] + strconv.FormatInt(int64(req.SilenceID), 10) + path[9+11:]

	request, err := a.client.prepareRequest("DELETE", a.client.cfg.BasePath+path, contentType, body)
	if err != nil {
		return
	}
	resp, err = a.client.Do(ctx, request)
	if err != nil {
		return
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode == http.StatusOK {
		return
	} else if resp.StatusCode >= 300 {
		err = errors.New(string(respBody))
	}

	return
}

// (GET /silence/{silenceID})
func (a *SilenceAPI) GetSilence(ctx context.Context, req *GetSilenceRequest) (ret *GettableSilence, resp *http.Response, err error) {
	var (
		contentType string
		body        any
	)
	path := "/silence/{silenceID}"
	path = path[:9] + strconv.FormatInt(int64(req.SilenceID), 10) + path[9+11:]

	request, err := a.client.prepareRequest("GET", a.client.cfg.BasePath+path, contentType, body)
	if err != nil {
		return
	}
	accept := selectHeaderAccept([]string{"application/json"})
	request.Header.Set("Accept", accept)
	resp, err = a.client.Do(ctx, request)
	if err != nil {
		return
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode == http.StatusOK {
		ret = new(GettableSilence)
		err = a.client.decode(respBody, ret, resp.Header.Get("Content-Type"))
		if err == nil {
			return
		}
	} else if resp.StatusCode >= 300 {
		err = errors.New(string(respBody))
	}

	return
}

// (GET /silences)
func (a *SilenceAPI) GetSilences(ctx context.Context, req *GetSilencesRequest) (ret GettableSilences, resp *http.Response, err error) {
	var (
		contentType string
		body        any
	)
	path := "/silences"
	queryParams := url.Values{}
	if req.Filter != nil {
		for _, v := range req.Filter {
			queryParams.Add("filter", v)
		}
	}

	request, err := a.client.prepareRequest("GET", a.client.cfg.BasePath+path, contentType, body)
	if err != nil {
		return
	}
	request.URL.RawQuery = queryParams.Encode()
	accept := selectHeaderAccept([]string{"application/json"})
	request.Header.Set("Accept", accept)
	resp, err = a.client.Do(ctx, request)
	if err != nil {
		return
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode == http.StatusOK {
		err = a.client.decode(respBody, &ret, resp.Header.Get("Content-Type"))
		if err == nil {
			return
		}
	} else if resp.StatusCode >= 300 {
		err = errors.New(string(respBody))
	}

	return
}

// (POST /silences)
func (a *SilenceAPI) PostSilences(ctx context.Context, req *PostSilencesRequest) (ret *PostSilencesResponse, resp *http.Response, err error) {
	var (
		contentType string
		body        any
	)
	path := "/silences"
	contentType = selectHeaderContentType([]string{"application/json"})
	body = req

	request, err := a.client.prepareRequest("POST", a.client.cfg.BasePath+path, contentType, body)
	if err != nil {
		return
	}
	accept := selectHeaderAccept([]string{"application/json"})
	request.Header.Set("Accept", accept)
	resp, err = a.client.Do(ctx, request)
	if err != nil {
		return
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode == http.StatusOK {
		ret = new(PostSilencesResponse)
		err = a.client.decode(respBody, ret, resp.Header.Get("Content-Type"))
		if err == nil {
			return
		}
	} else if resp.StatusCode >= 300 {
		err = errors.New(string(respBody))
	}

	return
}
