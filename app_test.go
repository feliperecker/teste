package core_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/go-bolo/core"
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
			got := core.NewApp(&core.DefaultAppOptions{})

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
			got := core.NewApp(&core.DefaultAppOptions{})
			err := got.Bootstrap()
			assert.Nil(t, err)

			approvals.VerifyJSONStruct(t, got)
		})
	}
}

func TestDefaultApp_AddPlugin(t *testing.T) {
	type fields struct {
		Plugins map[string]core.Plugin
	}
	type args struct {
		pluginName string
		p          core.Plugin
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &core.DefaultApp{
				Plugins: tt.fields.Plugins,
			}
			if err := app.AddPlugin(tt.args.pluginName, tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("DefaultApp.AddPlugin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultApp_GetPlugin(t *testing.T) {
	type fields struct {
		Plugins map[string]core.Plugin
	}
	type args struct {
		pluginName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wantP  core.Plugin
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &core.DefaultApp{
				Plugins: tt.fields.Plugins,
			}
			if gotP := app.GetPlugin(tt.args.pluginName); !reflect.DeepEqual(gotP, tt.wantP) {
				t.Errorf("DefaultApp.GetPlugin() = %v, want %v", gotP, tt.wantP)
			}
		})
	}
}

func TestDefaultApp_HasPlugin(t *testing.T) {
	type fields struct {
		Plugins map[string]core.Plugin
	}
	type args struct {
		pluginName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantHas bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &core.DefaultApp{
				Plugins: tt.fields.Plugins,
			}
			if gotHas := app.HasPlugin(tt.args.pluginName); gotHas != tt.wantHas {
				t.Errorf("DefaultApp.HasPlugin() = %v, want %v", gotHas, tt.wantHas)
			}
		})
	}
}

func TestRequestFlow(t *testing.T) {
	type fields struct {
		Plugins map[string]core.Plugin
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
		expectedError *core.HTTPError
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
			expectedError: &core.HTTPError{
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
			expectedError: &core.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  "oi2",
				Internal: errors.New("oi2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := GetTestApp()
			err := app.AddPlugin("example", &URLShortenerPlugin{})
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

			core.CtxSetAccept(c, tt.args.accept)

			rHandler := app.BindRoute("example_get", &core.Route{
				Method:   http.MethodGet,
				Path:     "/",
				Action:   p.Controller.Query,
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
