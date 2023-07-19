package bolo_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	bolo "github.com/go-bolo/bolo"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "should return a valid default app with required data",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bolo.NewApp(&bolo.DefaultAppOptions{})

			approvals.VerifyJSONStruct(t, got)
		})
	}
}
func TestApp_Bootstrap(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "should return a valid default app with required data",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bolo.NewApp(&bolo.DefaultAppOptions{})
			err := got.Bootstrap()
			assert.Nil(t, err)

			approvals.VerifyJSONStruct(t, got)
		})
	}
}

func TestRequestFlow(t *testing.T) {
	type fields struct {
		Plugins map[string]bolo.Plugin
	}
	type args struct {
		accept      string
		template    string
		queryParams string
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantHas       bool
		expectedError *bolo.HTTPError
	}{
		{
			name: "should run a action with success",
			args: args{
				accept: "application/json",
			},
		},
		{
			name: "should return a html page with success",
			args: args{
				accept:   "text/html",
				template: "urls/example",
			},
		},
		{
			name: "should return template not found error page",
			args: args{
				accept:   "text/html",
				template: "urls/unknown",
			},
		},
		{
			name: "should return a custom 500 error",
			args: args{
				accept:      "text/html",
				template:    "urls/example",
				queryParams: "errorToReturn=oi",
			},
			expectedError: &bolo.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  "",
				Internal: errors.New(""),
			},
		},
		{
			name: "should return a custom 500 error with message",
			args: args{
				accept:      "text/html",
				template:    "urls/example",
				queryParams: "errorToReturn=oi&errorMessage=oi2",
			},
			expectedError: &bolo.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  "oi2",
				Internal: errors.New("oi2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := GetTestApp()
			err := app.AddPlugin(&URLShortenerPlugin{Name: "example"})
			assert.Nil(t, err)

			err = app.Bootstrap()
			assert.Nil(t, err)

			p := app.GetPlugin("example").(*URLShortenerPlugin)

			e := app.GetRouter()
			req, err := http.NewRequest(http.MethodGet, "/?"+tt.args.queryParams, nil)
			if err != nil {
				t.Errorf("TestCreateOneHandler error: %v", err)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("app", app)

			bolo.SetAcceptCtx(c, tt.args.accept)

			rHandler := app.BindRoute("example_get", &bolo.Route{
				Method:   http.MethodGet,
				Path:     "/",
				Action:   p.Controller.Find,
				Template: tt.args.template,
			})

			err = rHandler(c)
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				return
			}

			assert.Nil(t, err)

			switch tt.args.accept {
			case "application/json":
				approvals.VerifyJSONBytes(t, rec.Body.Bytes())
			case "text/html":
				approvals.VerifyString(t, rec.Body.String())
			}
		})
	}
}

func TestRequest_CRUD(t *testing.T) {
	app := GetTestApp()
	err := app.AddPlugin(&URLShortenerPlugin{Name: "example"})
	assert.Nil(t, err)
	err = app.Bootstrap()
	assert.Nil(t, err)
	err = app.SyncDB()
	assert.Nil(t, err)

	app.GetAcl().SetDisabled(true)

	savedRecord1 := URLModel{
		Title: "Google",
		Path:  "http://www.google.com",
	}
	savedRecord1.Save(app)

	savedRecord2 := URLModel{
		Title: "Bing",
		Path:  "http://www.bing.com",
	}
	savedRecord2.Save(app)

	assert.Nil(t, err)

	type fields struct {
		Plugins map[string]bolo.Plugin
	}
	type args struct {
		accept      string
		queryParams string
		data        io.Reader
		url         string
		method      string
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantHas        bool
		expectedStatus int
		expectedError  *bolo.HTTPError
	}{
		{
			name: "should run a action with success",
			args: args{
				method: http.MethodGet,
				url:    "/urls",
				accept: "application/json",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "should return a html page with success",
			args: args{
				method: http.MethodGet,
				url:    "/urls",
				accept: "text/html",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "findOne should return one record",
			args: args{
				method: http.MethodGet,
				url:    "/urls/" + savedRecord1.GetID(),
				accept: "text/html",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "findOne should return 404 with invalid id",
			args: args{
				method: http.MethodGet,
				url:    "/urls/1111111111",
				accept: "text/html",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "JSON findOne should return 404 with invalid id",
			args: args{
				method: http.MethodGet,
				url:    "/urls/1111111111",
				accept: "application/json",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "JSON create should create a new record",
			args: args{
				method: http.MethodPost,
				url:    "/api/v1/urls",
				accept: "application/json",
				data:   strings.NewReader(`{"url":{"title":"example","path":"http://www.example.com"}}`),
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "JSON get count",
			args: args{
				method: http.MethodGet,
				url:    "/api/v1/urls-count",
				accept: "application/json",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "JSON should run update action",
			args: args{
				method: http.MethodPost,
				url:    "/api/v1/urls/1",
				accept: "application/json",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "JSON should run delete action",
			args: args{
				method: http.MethodDelete,
				url:    "/api/v1/urls/1",
				accept: "application/json",
			},
			expectedStatus: http.StatusNoContent,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := app.GetRouter()

			req := httptest.NewRequest(tt.args.method, tt.args.url, tt.args.data)
			req.Header.Set(echo.HeaderAccept, tt.args.accept)
			req.Header.Set(echo.HeaderContentType, "application/json")

			rec := httptest.NewRecorder() // run the request:
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			switch tt.args.accept {
			case "application/json":
				approvals.VerifyJSONBytes(t, rec.Body.Bytes())
			case "text/html":
				approvals.VerifyString(t, rec.Body.String())
			}
		})
	}
}
