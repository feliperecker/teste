package core_test

import (
	"github.com/go-bolo/core"
	"github.com/labstack/echo/v4"
)

// Mocks:

// URLShortener is a plugin that shortens URLs.
type URLShortenerPlugin struct {
	App        core.App `json:"-"`
	Controller core.Controller
}

// Init initializes the plugin.
func (p *URLShortenerPlugin) Init(app core.App) error {
	p.Controller = &URLController{
		// App: app,
	}

	// app.SetRoute(&core.Route{
	// 	Method: http.MethodGet,
	// 	Action: p.Controller.Query,
	// })

	return nil
}

// GetName returns the name of the plugin.
func (p *URLShortenerPlugin) GetName() string {
	return "URLShortenerPlugin"
}

// SetName sets the name of the plugin.
func (p *URLShortenerPlugin) SetName(name string) error {
	return nil
}

type URLController struct {
	App core.App `json:"-"`
}

func (ctl *URLController) Query(c echo.Context) (core.Response, error) {

	data := struct {
		Name string `json:"name"`
	}{
		Name: "oi",
	}

	r := core.DefaultResponse{
		Data: data,
	}

	return &r, nil
}
