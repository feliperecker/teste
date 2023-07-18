package bolo

import (
	"bytes"
	"html/template"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type ResponseMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

var (
	ResponseMessageKey = "responseMessage"
)

func SetResponseMessage(c echo.Context, key string, message *ResponseMessage) error {
	messages, err := GetResponseMessages(c)
	if err != nil {
		return err
	}

	if key == "" {
		key = uuid.New().String()
	}

	messages[key] = message
	c.Set(ResponseMessageKey, messages)

	return nil
}

func GetResponseMessages(c echo.Context) (map[string]*ResponseMessage, error) {
	iMessages := c.Get(ResponseMessageKey)

	switch ms := iMessages.(type) {
	case map[string]*ResponseMessage:
		return ms, nil
	}

	return map[string]*ResponseMessage{}, nil
}

func ResponseMessagesRender(c echo.Context, tpl string) template.HTML {
	app := c.Get("app").(App)
	l := app.GetLogger()

	html := ""

	messages, err := GetResponseMessages(c)
	if err != nil {
		return template.HTML(html)
	}

	if tpl == "" {
		tpl = "blocks/response/messages"
	}

	if app.HasTemplate(tpl) {
		var htmlBuffer bytes.Buffer
		err := app.RenderTemplate(&htmlBuffer, GetTheme(c), tpl, messages)
		if err != nil {
			l.Error("ResponseMessageRender error on render template", zap.Error(err), zap.String("template", tpl))
			return template.HTML(html)
		}

		html = htmlBuffer.String()
	}

	for _, m := range messages {
		html += `<div>` + m.Type + ": " + m.Message + `</div>'`
	}

	return template.HTML(html)
}
