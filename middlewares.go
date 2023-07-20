package bolo

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// BindMiddlewares - Bind middlewares in order
func BindMiddlewares(app App, p Plugin) {
	app.GetLogger().Debug("BindMiddlewares ", zap.String("plugin", p.GetName()))

	goEnv := app.GetConfiguration().Get(ENV_VARIABLE_NAME)

	router := app.GetRouter()
	router.Pre(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}))

	router.Pre(AcceptResolverMiddleware(app))

	router.Use(middleware.Gzip())
	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: app.GetConfiguration().GetBoolF(CORS_ALLOW_CREDENTIALS, true),
		MaxAge:           app.GetConfiguration().GetIntF(CORS_MAX_AGE, 18000), // secconds
	}))

	if goEnv == "dev" {
		router.Debug = true
	}
}

func AcceptResolverMiddleware(app App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			acceptType := NegotiateContentType(c.Request(), app.GetContentTypes(), app.GetDefaultContentType())
			SetAccept(c, acceptType)

			return next(c)
		}
	}
}
