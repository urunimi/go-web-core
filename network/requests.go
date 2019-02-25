package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/urunimi/go-web-core/core"
)

type Request struct {
	Method  string
	Url     string
	Params  *url.Values
	Header  *http.Header
	Timeout time.Duration

	httpRequest *http.Request
	client      *http.Client
}

func (req *Request) newHttpRequest() (*http.Request, error) {
	if req.Method == "" {
		req.Method = http.MethodGet
	}
	encodedParams := ""
	if req.Params != nil {
		encodedParams = req.Params.Encode()
	}
	var err error
	var httpRequest *http.Request = nil
	switch req.Method {
	case http.MethodGet:
		httpRequest, err = http.NewRequest(req.Method, req.Url, nil)
		httpRequest.URL.RawQuery = req.Params.Encode()
		if req.Header != nil {
			httpRequest.Header = *req.Header
		}
	case http.MethodPost:
		fallthrough
	case http.MethodDelete:
		fallthrough
	case http.MethodPut:
		httpRequest, err = http.NewRequest(req.Method, req.Url, bytes.NewBufferString(encodedParams))
		if req.Header != nil {
			httpRequest.Header = *req.Header
		}
		httpRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		httpRequest.Header.Add("Content-Length", strconv.Itoa(len(encodedParams)))
	default:
		err = errors.New("Not Yet Implemented")
	}

	return httpRequest, err
}

func (req *Request) request() (*http.Response, error) {
	if req.client == nil || req.httpRequest == nil {
		req.Build()
	}
	return req.client.Do(req.httpRequest)
}

func (req *Request) parseResponse(body io.ReadCloser, target interface{}) error {
	decoder := json.NewDecoder(body)
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		//if _, ok := err.(json.SyntaxError); ok {
		//	buf := new(bytes.Buffer)
		//	buf.ReadFrom(body)
		//	core.Logger().Panicf("Response Body cannot be parsed. Error - %s Body - %s", err, buf.String())
		//}
		return err
	}
	return nil
}

func (req *Request) Build() *Request {
	var err error
	req.httpRequest, err = req.newHttpRequest()
	if err != nil {
		core.Logger().Panic(err)
	}
	req.client = http.DefaultClient
	if req.Timeout > 0 {
		req.client.Timeout = req.Timeout
	}
	return req
}

func (req *Request) GetHttpRequest() *http.Request {
	if req.client == nil {
		req.Build()
	}
	return req.httpRequest
}

func (req *Request) GetResponse(target interface{}) (int, error) {
	httpResponse, err := req.request()
	if err != nil {
		return http.StatusInternalServerError, err
	} else if httpResponse == nil {
		return http.StatusInternalServerError, errors.New("Response is nil")
	} else if httpResponse.StatusCode >= http.StatusInternalServerError {
		return httpResponse.StatusCode, errors.New("Internal Server Error")
	} else if httpResponse.StatusCode >= http.StatusBadRequest {
		return httpResponse.StatusCode, errors.New("Bad Request Error")
	}

	defer httpResponse.Body.Close()
	if err := req.parseResponse(httpResponse.Body, target); err != nil {
		return httpResponse.StatusCode, err
	}
	return httpResponse.StatusCode, nil
}
