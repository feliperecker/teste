package core

import "github.com/labstack/echo/v4"

type Controller interface {
	Query(c echo.Context) (Response, error)
	// Create(c echo.Context) error
	// Count(c echo.Context) error
	// FindOne(c echo.Context) error
	// Update(c echo.Context) error
	// Delete(c echo.Context) error
}
