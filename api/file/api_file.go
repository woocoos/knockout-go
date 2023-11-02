// Code generated by woco, DO NOT EDIT.

package file

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
)

type FileAPI api

// (DELETE /files/{fileId})
func (a *FileAPI) DeleteFile(ctx context.Context, req *DeleteFileRequest) (resp *http.Response, err error) {
	var (
		contentType string
		body        any
	)
	path := "/files/{fileId}"
	path = path[:7] + req.FileId + path[7+8:]

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

// (GET /files/{fileId})
func (a *FileAPI) GetFile(ctx context.Context, req *GetFileRequest) (ret *FileInfo, resp *http.Response, err error) {
	var (
		contentType string
		body        any
	)
	path := "/files/{fileId}"
	path = path[:7] + req.FileId + path[7+8:]

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
		ret = new(FileInfo)
		err = a.client.decode(respBody, ret, resp.Header.Get("Content-Type"))
		if err == nil {
			return
		}
	} else if resp.StatusCode >= 300 {
		err = errors.New(string(respBody))
	}

	return
}

// (GET /files/{fileId}/raw)
func (a *FileAPI) GetFileRaw(ctx context.Context, req *GetFileRawRequest) (ret []byte, resp *http.Response, err error) {
	var (
		contentType string
		body        any
	)
	path := "/files/{fileId}/raw"
	path = path[:7] + req.FileId + path[7+8:]

	request, err := a.client.prepareRequest("GET", a.client.cfg.BasePath+path, contentType, body)
	if err != nil {
		return
	}
	accept := selectHeaderAccept([]string{"application/octet-stream"})
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

// (POST /files/report-ref-count)
func (a *FileAPI) ReportRefCount(ctx context.Context, req *ReportRefCountRequest) (ret bool, resp *http.Response, err error) {
	var (
		contentType string
		body        any
	)
	path := "/files/report-ref-count"
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
		err = a.client.decode(respBody, &ret, resp.Header.Get("Content-Type"))
		if err == nil {
			return
		}
	} else if resp.StatusCode >= 300 {
		err = errors.New(string(respBody))
	}

	return
}

// (POST /files)
func (a *FileAPI) UploadFile(ctx context.Context, req *UploadFileRequest) (ret string, resp *http.Response, err error) {
	var (
		contentType string
		body        any
	)
	path := "/files"
	contentType = selectHeaderContentType([]string{"multipart/form-data"})
	forms := url.Values{}
	forms.Add("bucket", req.Bucket)
	forms.Add("key", req.Key)
	body = forms

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
		err = a.client.decode(respBody, &ret, resp.Header.Get("Content-Type"))
		if err == nil {
			return
		}
	} else if resp.StatusCode >= 300 {
		err = errors.New(string(respBody))
	}

	return
}

// (POST /files/upload-info)
func (a *FileAPI) UploadFileInfo(ctx context.Context, req *UploadFileInfoRequest) (ret string, resp *http.Response, err error) {
	var (
		contentType string
		body        any
	)
	path := "/files/upload-info"
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
		err = a.client.decode(respBody, &ret, resp.Header.Get("Content-Type"))
		if err == nil {
			return
		}
	} else if resp.StatusCode >= 300 {
		err = errors.New(string(respBody))
	}

	return
}
