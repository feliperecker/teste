package core_test

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-bolo/core"
	"github.com/gookit/event"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
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

	app.SetModel("url", &URLModel{})

	app.GetEvents().On("bindRoutes", event.ListenerFunc(func(e event.Event) error {
		return p.BindRoutes(app)
	}), event.Normal)

	// app.GetEvents().On("bootstrap", event.ListenerFunc(func(e event.Event) error {
	// 	return p.Bootstrap(app)
	// }), event.Normal)

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

func (p *URLShortenerPlugin) BindRoutes(app core.App) error {
	ctl := p.Controller

	app.SetRoute("urls_query", &core.Route{
		Method:   http.MethodGet,
		Path:     "/urls",
		Action:   ctl.Find,
		Template: "urls/find",
	})

	app.SetRoute("urls_findOne", &core.Route{
		Method:   http.MethodGet,
		Path:     "/urls/:id",
		Action:   ctl.FindOne,
		Template: "urls/findOne",
	})

	// app.SetResource("urls", ctl, "/api/v1/urls")
	app.SetResource(&core.Resource{
		Name:       "urls",
		Prefix:     "/api/v1",
		Path:       "/urls",
		Controller: ctl,
		Model:      &URLModel{},
	})

	return nil
}

type JSONResponse struct {
	core.BaseListReponse
	URLs *[]*URLModel `json:"url"`
}

type CountJSONResponse struct {
	core.BaseMetaResponse
}

type FindOneJSONResponse struct {
	URL *URLModel `json:"url"`
}

type BodyRequest struct {
	URL *URLModel `json:"url"`
}

type URLController struct {
	App core.App `json:"-"`
}

func (ctl *URLController) Find(c echo.Context) (core.Response, error) {
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

func (ctl *URLController) Create(c echo.Context) (core.Response, error) {
	var err error

	app := core.GetAppCtx(c)
	acl := app.GetAcl()
	route := core.GetRouteCtx(c)

	can := acl.Can(route.Permission, core.GetRolesCtx(c))
	if !can {
		return nil, &core.HTTPError{
			Code:    http.StatusForbidden,
			Message: "Forbidden",
		}
	}

	var body BodyRequest

	if err := c.Bind(&body); err != nil {
		if er, ok := err.(*echo.HTTPError); ok {

			return nil, &core.HTTPError{
				Code:     er.Code,
				Message:  er.Message,
				Internal: er.Internal,
			}
		}

		return nil, &core.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  "Invalid body data",
			Internal: fmt.Errorf("urls.Create error on parse body: %w", err),
		}
	}

	record := body.URL
	record.ID = 0

	if core.IsAuthenticatedCtx(c) {
		creatorID := core.GetAuthenticatedUserCtx(c).GetID()
		record.CreatorID = &creatorID
	}

	if err := c.Validate(record); err != nil {
		if _, ok := err.(*echo.HTTPError); ok {
			return nil, err
		}
		return nil, &core.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  "Error on validate data",
			Internal: err,
		}
	}

	err = record.Save(app)
	if err != nil {
		return nil, err
	}

	resp := FindOneJSONResponse{
		URL: record,
	}

	r := core.DefaultResponse{
		Status: http.StatusCreated,
		Data:   resp,
	}

	return &r, nil
}

func (ctl *URLController) FindOne(c echo.Context) (core.Response, error) {
	app := core.GetAppCtx(c)
	id := c.Param("id")

	record, err := FindOneURL(app, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &core.HTTPError{
				Code:     http.StatusNotFound,
				Message:  "Not found",
				Internal: err,
			}
		}

		return nil, &core.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  "Error on FindOneURL",
			Internal: err,
		}
	}

	r := core.DefaultResponse{
		Status: http.StatusOK,
		Data:   record,
	}

	return &r, nil
}

func (ctl *URLController) Count(c echo.Context) (core.Response, error) {
	r := core.DefaultResponse{
		Data: CountJSONResponse{
			BaseMetaResponse: core.BaseMetaResponse{
				Count: 90,
			},
		},
	}

	return &r, nil
}

func (ctl *URLController) Update(c echo.Context) (core.Response, error) {
	panic("TODO!")
}

func (ctl *URLController) Delete(c echo.Context) (core.Response, error) {
	panic("TODO!")
}
