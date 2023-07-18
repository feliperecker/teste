package core

import (
	"fmt"

	"github.com/go-bolo/core/pagination"
	"github.com/go-bolo/query_parser_to_db"
	"github.com/google/uuid"
	"github.com/gookit/event"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func GetAppCtx(c echo.Context) App {
	return c.Get("app").(App)
}

// SetDefaultValuesCtx - Ran at request start as a middleware to set all core values in echo context.
func SetDefaultValuesCtx(c echo.Context, app App) error {
	cfg := app.GetConfiguration()
	port := cfg.GetF("PORT", "8080")
	protocol := cfg.GetF("PROTOCOL", "http")
	domain := cfg.GetF("DOMAIN", "localhost")

	c.Set("app", app)
	c.Set("logger", app.GetLogger().With(zap.String("RID", uuid.New().String())))
	c.Set("theme", cfg.GetF("THEME", app.GetTheme()))
	c.Set("base_url", cfg.GetF("BASE_URL", protocol+"://"+domain+":"+port))

	SetPagerCtx(c, pagination.NewPager())
	SetQueryParserCtx(c, query_parser_to_db.NewQuery(50))
	SetAcceptCtx(c, app.GetDefaultContentType())
	SetAuthenticatedUserCtx(c, nil)
	SetRolesCtx(c, []string{})
	SetMetadataCtx(c, NewMetadata())

	err, _ := app.GetEvents().Fire("set-default-request-context-values", event.M{"app": app, "context": c})
	if err != nil {
		return fmt.Errorf("error on trigger set-default-request-context-values event: %w", err)
	}

	return nil
}

// GetRouteCtx - Returns the current request related route configuration
func GetRouteCtx(c echo.Context) *Route {
	return c.Get("route").(*Route)
}

// GetAcceptCtx - Returns the content type for the response.
// Accept is used to determine the response format or the default App configuration
func GetAcceptCtx(c echo.Context) string {
	accept := c.Get("accept")
	if accept == nil {
		return c.Get("app").(App).GetDefaultContentType()
	}

	return accept.(string)
}

// SetAcceptCtx - Accept used to determine the response format.
func SetAcceptCtx(c echo.Context, accept string) {
	c.Set("accept", accept)
}

func GetAuthenticatedUserCtx(c echo.Context) User {
	user := c.Get("user")
	if user == nil {
		return nil
	}

	return user.(User)
}

func SetAuthenticatedUserCtx(c echo.Context, user User) error {
	c.Set("user", user)
	return nil
}

// GetRolesCtx - Returns the current request related route configuration
func GetRolesCtx(c echo.Context) []string {
	roles := c.Get("roles")
	if roles == nil {
		return []string{}
	}

	return roles.([]string)
}

func SetRolesCtx(c echo.Context, roles []string) error {
	c.Set("roles", roles)
	return nil
}

func IsAuthenticatedCtx(c echo.Context) bool {
	return GetAuthenticatedUserCtx(c) != nil
}

func GetQueryParserCtx(c echo.Context) query_parser_to_db.QueryInterface {
	queryParser := c.Get("query_parser")
	if queryParser == nil {
		return nil
	}

	return queryParser.(query_parser_to_db.QueryInterface)
}

func SetQueryParserCtx(c echo.Context, queryParser query_parser_to_db.QueryInterface) error {
	c.Set("query_parser", queryParser)
	return nil
}

func GetMetadataCtx(c echo.Context) Metadata {
	metadata := c.Get("metadata")
	if metadata == nil {
		return NewMetadata()
	}

	return metadata.(Metadata)
}

func SetMetadataCtx(c echo.Context, metadata Metadata) error {
	c.Set("metadata", metadata)
	return nil
}

func GetLoggerCtx(c echo.Context) *zap.Logger {
	return c.Get("logger").(*zap.Logger)
}

func GetPagerCtx(c echo.Context) *pagination.Pager {
	return c.Get("pager").(*pagination.Pager)
}

// SetPagerCtx - Set the pagination pager object in the echo context
func SetPagerCtx(c echo.Context, pager *pagination.Pager) error {
	c.Set("pager", pager)
	return nil
}
