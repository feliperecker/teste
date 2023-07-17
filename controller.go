package core

import "github.com/labstack/echo/v4"

type Controller interface {
	Find(c echo.Context) (Response, error)
	Create(c echo.Context) (Response, error)
	Count(c echo.Context) (Response, error)
	FindOne(c echo.Context) (Response, error)
	Update(c echo.Context) (Response, error)
	Delete(c echo.Context) (Response, error)
}
