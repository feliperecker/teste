package bolo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-bolo/bolo/configuration"
	"go.uber.org/zap"
)

// CustomHTTPClient - Custom http client required to make requests testable
type CustomHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	HttpClient CustomHTTPClient
)

func HttpClientInit() {
	httpClientTimeout := configuration.GetInt64Env("HTTP_CLIENT_TIMEOUT", 120)

	timeout := time.Second * time.Duration(httpClientTimeout)
	HttpClient = &http.Client{Timeout: timeout}
}

// DownloadFile - Download one file
func DownloadFile(app App, url string, dest *os.File, headers http.Header) (bool, error) {
	l := app.GetLogger().With(zap.String("func", "DownloadFile"))
	l.Debug("will download", zap.String("url", url), zap.String("dest", dest.Name()))

	var err error

	defer dest.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header = headers

	res, err := HttpClient.Do(req)
	if err != nil {
		l.Error("error on request", zap.Error(err), zap.String("url", url))
		return false, err
	}

	defer res.Body.Close()

	_, err = io.Copy(dest, res.Body)
	if err != nil {
		return false, err
	}

	l.Error("error on copy body", zap.Error(err), zap.String("url", url), zap.String("dest", dest.Name()))
	return true, err
}

// Get - Start a Get response and returns the http.Response without parse data.
func Get(url string, headers http.Header) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header = headers

	return HttpClient.Do(req)
}

func GetPageHTML(app App, url string, headers http.Header) (string, error) {
	l := app.GetLogger().With(zap.String("func", "GetPageHTML"))

	resp, err := Get(url, headers)
	if err != nil {
		l.Error("error on request", zap.Error(err), zap.String("url", url), zap.Any("headers", headers))
		return "", fmt.Errorf("error on request: %w", err)
	}

	defer resp.Body.Close()

	rdrBody := io.Reader(resp.Body)
	bodyBytes, err := io.ReadAll(rdrBody)
	if err != nil {
		l.Error("error on read body data", zap.Error(err), zap.String("url", url), zap.Any("headers", headers))
		return "", fmt.Errorf("error on read body data: %w", err)
	}

	return string(bodyBytes), nil
}

// Post sends a post request to the URL with the body
func Post(url string, body interface{}, headers http.Header) (*http.Response, error) {
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}
	request.Header = headers
	return HttpClient.Do(request)
}

// PostFormURLEncoded - Send a post request with form url encoded request body
func PostFormURLEncoded(url string, body url.Values, target interface{}) error {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(bodyBytes, target)
}
