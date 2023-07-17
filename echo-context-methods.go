package core

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/gookit/event"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func GetAppCtx(c echo.Context) App {
	return c.Get("app").(App)
}

func SetDefaultValuesCtx(c echo.Context, app App) error {
	c.Set("app", app)
	c.Set("logger", app.GetLogger().With(zap.String("RID", uuid.New().String())))

	// SetAcceptCtx(c, app.GetDefaultContentType())
	SetAuthenticatedUserCtx(c, nil)
	SetRolesCtx(c, []string{})
	SetMetadataCtx(c, NewMetadata())

	err, _ := app.GetEvents().Fire("set-default-request-context-values", event.M{"app": app, "context": c})
	if err != nil {
		return fmt.Errorf("error on trigger set-default-request-context-values event: %w", err)
	}

	return nil
}

func GetRouteCtx(c echo.Context) *Route {
	return c.Get("route").(*Route)
}

// SetAccept - Accept used to determine the response format.
func SetAcceptCtx(c echo.Context, accept string) {
	c.Set("accept", accept)
}

// GetAccept - Accept is used to determine the response format.
func GetAcceptCtx(c echo.Context) string {
	accept := c.Get("accept")
	if accept == nil {
		return c.Get("app").(App).GetDefaultContentType()
	}

	return accept.(string)
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
