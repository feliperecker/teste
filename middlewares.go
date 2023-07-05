package core

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

// BindMiddlewares - Bind middlewares in order
func BindMiddlewares(app App, p Plugin) {
	logrus.Debug("catu.BindMiddlewares " + p.GetName())

	goEnv := app.GetConfiguration().Get("GO_ENV")

	router := app.GetRouter()
	router.Pre(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}))

	router.Use(middleware.Gzip())
	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: app.GetConfiguration().GetBoolF("CORS_ALLOW_CREDENTIALS", true),
		MaxAge:           app.GetConfiguration().GetIntF("CORS_MAX_AGE", 18000), // secconds
	}))

	if goEnv == "dev" {
		router.Debug = true
	}
}

func isPublicRoute(url string) bool {
	return strings.HasPrefix(url, "/health") || strings.HasPrefix(url, "/public")
}
