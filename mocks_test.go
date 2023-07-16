package core_test

import (
	"errors"
	"strconv"

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
		App: app,
	}

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

	if c.QueryParam("errorToReturn") != "" {
		eCode := c.QueryParam("errorCode")
		eMessage := c.QueryParam("errorMessage")
		eCodeInt, _ := strconv.Atoi(eCode)
		if eCodeInt == 0 {
			eCodeInt = 500
		}

		return nil, &core.HTTPError{
			Code:     eCodeInt,
			Message:  eMessage,
			Internal: errors.New(eMessage),
		}
	}

	r := core.DefaultResponse{
		Data: data,
	}

	return &r, nil
}
