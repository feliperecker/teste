package bolo

import (
	"github.com/labstack/echo/v4"
)

func SetDefaultResponseFormatters(app App) error {
	app.SetResponseFormatter("application/json", DefaultJSONFormatter)
	app.SetResponseFormatter("text/html", DefaultHTMLFormatter)
	return nil
}

func DefaultJSONFormatter(app App, c echo.Context, r *Route, resp Response) error {
	return c.JSON(resp.GetStatusCode(), resp.GetData())
}

func DefaultHTMLFormatter(app App, c echo.Context, r *Route, resp Response) error {
	template := app.GetTemplateCtx(c, r)

	return MinifiAndRender(resp.GetStatusCode(), template, &TemplateCTX{
		Ctx:  c,
		Data: resp.GetData(),
	}, c)
}
