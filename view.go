package bolo

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-bolo/bolo/pagination"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Metadata interface {
	Get(key string) string
	Set(key string, value string)
	Remove(key string)
	GetAll() map[string]string
}

func NewMetadata() *MetadataDefault {
	return &MetadataDefault{
		Data: make(map[string]string),
	}
}

type MetadataDefault struct {
	Data map[string]string
}

func (m *MetadataDefault) Get(key string) string {
	return m.Data[key]
}

func (m *MetadataDefault) Set(key string, value string) {
	m.Data[key] = value
}

func (m *MetadataDefault) Remove(key string) {
	delete(m.Data, key)
}

func (m *MetadataDefault) GetAll() map[string]string {
	return m.Data
}

type TemplateCTX struct {
	Ctx     echo.Context
	Data    any
	Content template.HTML
}

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	app := c.Get("app").(App)
	l := app.GetLogger()

	switch v := data.(type) {
	case int:
		// v is an int here, so e.g. v + 1 is possible.
		fmt.Printf("Integer: %v", v)
	case float64:
		// v is a float64 here, so e.g. v + 1.0 is possible.
		fmt.Printf("Float64: %v", v)
	case string:
		// v is a string here, so e.g. v + " Yeah!" is possible.
		fmt.Printf("String: %v", v)
	default:
		htmlContext := data.(*TemplateCTX)
		theme := GetTheme(c)
		layout := GetLayout(c)

		if app.GetEnv() == "development" {
			l.Debug("Render", zap.Any("htmlContext", htmlContext), zap.String("name", name))
		}

		var contentBuffer bytes.Buffer
		err := app.RenderTemplate(&contentBuffer, theme, name, htmlContext)
		if err != nil {
			if strings.Contains(err.Error(), "is undefined") {
				l.Error("Render error: template not found", zap.Error(err), zap.String("name", name), zap.String("theme", theme))
				return c.String(http.StatusNotImplemented, "Template "+name+" not found: theme="+theme+" layout="+layout)
			}

			l.Error("Render error: on render template", zap.Error(err), zap.String("name", name), zap.String("theme", theme))

			return err
		}

		htmlContext.Content = template.HTML(contentBuffer.String())

		var layoutBuffer bytes.Buffer
		err = app.RenderTemplate(&layoutBuffer, theme, layout, htmlContext)
		if err != nil {
			l.Error("Render error on render layout", zap.Error(err), zap.String("name", name), zap.String("layout", layout), zap.String("theme", theme))
			return err
		}

		htmlContext.Content = template.HTML(layoutBuffer.String())

		err = app.RenderTemplate(w, theme, "html", htmlContext)
		return err
	}

	return nil
}

func findAndParseTemplates(rootDir string, funcMap template.FuncMap) (*template.Template, error) {
	cleanRoot := filepath.Clean(rootDir)
	pfx := len(cleanRoot) + 1
	root := template.New("")

	err := filepath.Walk(cleanRoot, func(path string, info os.FileInfo, e1 error) error {
		if info != nil && !info.IsDir() && strings.HasSuffix(path, ".html") {
			if e1 != nil {
				return e1
			}

			b, e2 := ioutil.ReadFile(path)
			if e2 != nil {
				return e2
			}

			name := path[pfx:]
			name = strings.Replace(name, ".html", "", 1)

			t := root.New(name).Funcs(funcMap)
			_, e2 = t.Parse(string(b))
			if e2 != nil {
				return e2
			}
		}

		return nil
	})

	return root, err
}

func renderPager(c echo.Context, r *pagination.Pager, queryString string) template.HTML {
	var htmlBuffer bytes.Buffer

	app := c.Get("app").(App)
	app.GetLogger().Debug("renderPager", zap.Any("pager", r), zap.String("queryString", queryString))

	if r.Count == 0 {
		return template.HTML("")
	}

	currentUrl := r.CurrentUrl
	queryParamsStr := ""

	if queryString != "" {
		queryParamsStr += "&" + queryString
	}

	pageCountFloat := float64(r.Count) / float64(r.Limit)
	pageCount := int64(math.Ceil(pageCountFloat))
	totalLinks := (r.MaxLinks * 2) + 1
	startInPage := int64(1)
	endInPage := pageCount

	if pageCount == 0 {
		return template.HTML("")
	}

	if totalLinks < pageCount {
		if r.MaxLinks+2 < r.Page {
			startInPage = r.Page - r.MaxLinks
			r.FirstPath = currentUrl + "?page=1" + queryParamsStr
			r.FirstNumber = "1"
			r.HasMoreBefore = true
		}

		if (r.MaxLinks + r.Page + 1) < pageCount {
			endInPage = r.MaxLinks + r.Page
			r.LastPath = currentUrl + "?page=" + strconv.FormatInt(pageCount, 10) + queryParamsStr
			r.LastNumber = strconv.FormatInt(pageCount, 10)
			r.HasMoreAfter = true
		}
	}

	// Each link
	for i := startInPage; i <= endInPage; i++ {
		number := strconv.FormatInt(i, 10)
		var link = pagination.Link{
			Path:   currentUrl + "?page=" + number + queryParamsStr,
			Number: number,
		}

		if i == r.Page {
			link.IsActive = true
		}

		r.Links = append(r.Links, link)
	}

	if r.Page > 1 {
		r.HasPrevius = true
		number := strconv.FormatInt(r.Page-1, 10)
		r.PreviusPath = currentUrl + "?page=" + number + queryParamsStr
		r.PreviusNumber = number
	}

	if r.Page < pageCount {
		r.HasNext = true
		number := strconv.FormatInt(r.Page+1, 10)
		r.NextPath = currentUrl + "?page=" + number + queryParamsStr
		r.NextNumber = number
	}

	return template.HTML(htmlBuffer.String())
}

func GetTheme(c echo.Context) string {
	t := c.Get("theme")
	if t != nil {
		return t.(string)
	}

	route := GetRoute(c)
	if route.Theme != "" {
		return route.Theme
	}

	app := GetApp(c)
	return app.GetTheme()
}

func SetTheme(c echo.Context, theme string) {
	c.Set("theme", theme)
}

func GetLayout(c echo.Context) string {
	l := c.Get("layout")
	if l != nil {
		return l.(string)
	}

	return "layouts/default"
}

func SetLayout(c echo.Context, theme string) {
	c.Set("layout", theme)
}

func SetCoreTemplateFunctions(app App) error {
	app.SetTemplateFunction("paginate", paginate)
	app.SetTemplateFunction("contentDates", contentDates)
	app.SetTemplateFunction("truncate", truncate(app))
	app.SetTemplateFunction("formatDecimalWithDots", formatDecimalWithDots)
	app.SetTemplateFunction("html", noEscapeHTML)
	app.SetTemplateFunction("currentDate", currentDate)
	app.SetTemplateFunction("responseMessagesRender", ResponseMessagesRender)
	app.SetTemplateFunction("partial", partial)

	return nil
}
