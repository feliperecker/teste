package core

import (
	"bytes"
	"html/template"

	"github.com/go-bolo/core/helpers"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func noEscapeHTML(str string) template.HTML {
	return template.HTML(str)
}

func paginate(c echo.Context, pager *Pager, queryString string) template.HTML {
	return renderPager(c, pager, queryString)
}

type ContentDates interface {
	GetTeaserDatesHTML(separator string) template.HTML
}

func contentDates(record ContentDates, separator string) template.HTML {
	return record.GetTeaserDatesHTML(separator)
}

func truncate(app App) func(text string, length int, ellipsis string) template.HTML {
	l := app.GetLogger()
	return func(text string, length int, ellipsis string) template.HTML {
		html, err := helpers.Truncate(text, length, ellipsis)
		if err != nil {
			l.Error("truncate error on truncate text", zap.Error(err), zap.String("text", text), zap.Int("length", length), zap.String("ellipsis", ellipsis))
		}
		return html
	}
}

func formatDecimalWithDots(value decimal.Decimal) string {
	return helpers.FormatDecimalWithDots(value)
}

func currentDate(format string) string {
	return helpers.FormatCurrencyDate(format)
}

func partial(name string, tplCtx TemplateCTX) template.HTML {
	var htmlBuffer bytes.Buffer
	app := tplCtx.Ctx.Get("app").(App)

	err := app.RenderTemplate(&htmlBuffer, GetTheme(tplCtx.Ctx), name, tplCtx)
	if err != nil {
		app.GetLogger().Error("Partial: error on render partial template", zap.Error(err), zap.String("partialName", name))
		return template.HTML("")
	}

	return template.HTML(htmlBuffer.String())
}
