package core

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/gookit/event"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func CtxSetDefaultValues(c echo.Context, app App) error {
	c.Set("app", app)
	c.Set("logger", app.GetLogger().With(zap.String("RID", uuid.New().String())))

	// CtxSetAccept(c, app.GetDefaultContentType())
	CtxSetAuthenticatedUser(c, nil)
	CtxSetRoles(c, []string{})
	CtxSetMetadata(c, NewMetadata())

	err, _ := app.GetEvents().Fire("set-default-request-context-values", event.M{"app": app, "context": c})
	if err != nil {
		return fmt.Errorf("error on trigger set-default-request-context-values event: %w", err)
	}

	return nil
}

// SetAccept - Accept used to determine the response format.
func CtxSetAccept(c echo.Context, accept string) {
	c.Set("accept", accept)
}

// GetAccept - Accept is used to determine the response format.
func CtxGetAccept(c echo.Context) string {
	accept := c.Get("accept")
	if accept == nil {
		return c.Get("app").(App).GetDefaultContentType()
	}

	return accept.(string)
}

func CtxGetAuthenticatedUser(c echo.Context) User {
	user := c.Get("user")
	if user == nil {
		return nil
	}

	return user.(User)
}

func CtxSetAuthenticatedUser(c echo.Context, user User) error {
	c.Set("user", user)
	return nil
}

func CtxGetRoles(c echo.Context) []string {
	roles := c.Get("roles")
	if roles == nil {
		return []string{}
	}

	return roles.([]string)
}

func CtxSetRoles(c echo.Context, roles []string) error {
	c.Set("roles", roles)
	return nil
}

func CtxIsAuthenticated(c echo.Context) bool {
	return CtxGetAuthenticatedUser(c) != nil
}

func CtxGetMetadata(c echo.Context) Metadata {
	metadata := c.Get("metadata")
	if metadata == nil {
		return NewMetadata()
	}

	return metadata.(Metadata)
}

func CtxSetMetadata(c echo.Context, metadata Metadata) error {
	c.Set("metadata", metadata)
	return nil
}

func CtxGetLogger(c echo.Context) zap.Logger {
	return c.Get("logger").(zap.Logger)
}
