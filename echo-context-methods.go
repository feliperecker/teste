package bolo

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-bolo/bolo/pagination"
	"github.com/go-bolo/query_parser_to_db"
	"github.com/google/uuid"
	"github.com/gookit/event"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func GetApp(c echo.Context) App {
	return c.Get("app").(App)
}

func NewContext(app App) echo.Context {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := app.GetRouter().NewContext(req, res)

	return c
}

// SetDefaultValuesCtx - Ran at request start as a middleware to set all core values in echo context.
func SetDefaultValues(c echo.Context, app App) error {
	cfg := app.GetConfiguration()
	port := cfg.GetF("PORT", "8080")
	protocol := cfg.GetF("PROTOCOL", "http")
	domain := cfg.GetF("DOMAIN", "localhost")

	c.Set("app", app)
	c.Set("logger", app.GetLogger().With(zap.String("RID", uuid.New().String())))
	c.Set("theme", cfg.GetF("THEME", app.GetTheme()))
	c.Set("base_url", cfg.GetF("BASE_URL", protocol+"://"+domain+":"+port))

	SetPager(c, pagination.NewPager())
	SetQueryParser(c, query_parser_to_db.NewQuery(50))
	SetAccept(c, app.GetDefaultContentType())
	SetAuthenticatedUser(c, nil)
	SetRoles(c, []string{})
	SetMetadata(c, NewMetadata())

	err, _ := app.GetEvents().Fire("set-default-request-context-values", event.M{"app": app, "context": c})
	if err != nil {
		return fmt.Errorf("error on trigger set-default-request-context-values event: %w", err)
	}

	return nil
}

// GetRouteCtx - Returns the current request related route configuration
func GetRoute(c echo.Context) *Route {
	return c.Get("route").(*Route)
}

// GetAcceptCtx - Returns the content type for the response.
// Accept is used to determine the response format or the default App configuration
func GetAccept(c echo.Context) string {
	accept := c.Get("accept")
	if accept == nil {
		return c.Get("app").(App).GetDefaultContentType()
	}

	return accept.(string)
}

// SetAcceptCtx - Accept used to determine the response format.
func SetAccept(c echo.Context, accept string) {
	c.Set("accept", accept)
}

func GetAuthenticatedUser(c echo.Context) User {
	user := c.Get("user")
	if user == nil {
		return nil
	}

	return user.(User)
}

func SetAuthenticatedUser(c echo.Context, user User) error {
	c.Set("user", user)
	return nil
}

// GetRolesCtx - Returns the current request related route configuration
func GetRoles(c echo.Context) []string {
	roles := c.Get("roles")
	if roles == nil {
		return []string{}
	}

	return roles.([]string)
}

func SetRoles(c echo.Context, roles []string) error {
	c.Set("roles", roles)
	return nil
}

// AddRoleCtx - Add a role to the current request context that is usualy used with access checks
func AddRole(c echo.Context, role string) error {
	roles := GetRoles(c)
	roles = append(roles, role)
	c.Set("roles", roles)
	return nil
}

func IsAuthenticated(c echo.Context) bool {
	return GetAuthenticatedUser(c) != nil
}

func Can(c echo.Context, permission string) bool {
	app := GetApp(c)
	roles := GetRoles(c)
	return app.GetAcl().Can(permission, roles)
}

func GetQueryParser(c echo.Context) query_parser_to_db.QueryInterface {
	queryParser := c.Get("query_parser")
	if queryParser == nil {
		return nil
	}

	return queryParser.(query_parser_to_db.QueryInterface)
}

func SetQueryParser(c echo.Context, queryParser query_parser_to_db.QueryInterface) error {
	c.Set("query_parser", queryParser)
	return nil
}

func GetMetadata(c echo.Context) Metadata {
	metadata := c.Get("metadata")
	if metadata == nil {
		return NewMetadata()
	}

	return metadata.(Metadata)
}

func SetMetadata(c echo.Context, metadata Metadata) error {
	c.Set("metadata", metadata)
	return nil
}

func GetLogger(c echo.Context) *zap.Logger {
	return c.Get("logger").(*zap.Logger)
}

func GetPager(c echo.Context) *pagination.Pager {
	return c.Get("pager").(*pagination.Pager)
}

// SetPagerCtx - Set the pagination pager object in the echo context
func SetPager(c echo.Context, pager *pagination.Pager) error {
	c.Set("pager", pager)
	return nil
}

// GetLimitCtx - Get the limit for the response list for list routes. Ex query
func GetLimit(c echo.Context) int {
	p := GetPager(c)
	return int(p.Limit)
}

// GetOffsetCtx - Calculate and returns the route request query offset
func GetOffset(c echo.Context) int {
	pager := GetPager(c)
	page := int(pager.Page)

	if page < 2 {
		return 0
	}

	limit := int(pager.Limit)
	return limit * (page - 1)
}

func GetBaseURL(c echo.Context) string {
	baseURL := c.Get("base_url")
	if baseURL == nil {
		return ""
	}

	return baseURL.(string)
}

func SetBaseURL(c echo.Context, baseURL string) error {
	c.Set("base_url", baseURL)
	return nil
}
