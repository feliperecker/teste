package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-bolo/core/configuration"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// CustomHTTPClient - Custom http client required to make requests testable
type CustomHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	HttpClient CustomHTTPClient
)

func Init() {
	httpClientTimeout := configuration.GetInt64Env("HTTP_CLIENT_TIMEOUT", 120)

	timeout := time.Second * time.Duration(httpClientTimeout)
	HttpClient = &http.Client{Timeout: timeout}
}

// DownloadFile - Download one file
func DownloadFile(url string, dest *os.File, headers http.Header) (bool, error) {
	logrus.WithFields(logrus.Fields{
		"url":  url,
		"dest": dest.Name(),
	}).Debug("DownloadFile will download")

	var err error

	defer dest.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header = headers

	res, err := HttpClient.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"url":     url,
			"headers": headers,
			"error":   err,
		}).Error("DownloadFile error")
		return false, err
	}

	defer res.Body.Close()

	_, err = io.Copy(dest, res.Body)
	if err != nil {
		return false, err
	}

	logrus.WithFields(logrus.Fields{
		"url":  url,
		"dest": dest.Name(),
	}).Debug("DownloadFile done download")

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

func GetPageHTML(url string, headers http.Header) (string, error) {
	resp, err := Get(url, headers)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"url":     url,
			"headers": headers,
			"error":   err,
		}).Error("GetPageHTML error")
		return "", err
	}

	defer resp.Body.Close()

	rdrBody := io.Reader(resp.Body)
	bodyBytes, err := ioutil.ReadAll(rdrBody)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": fmt.Sprintf("%+v\n", err),
		}).Debug("catu.GetPageHTML error")
		return "", errors.Wrap(err, "GetPageHTML error")
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

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(bodyBytes, target)
}
