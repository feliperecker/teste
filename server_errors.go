package bolo

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewHTTPError(code int, message interface{}, internal error) error {
	return &HTTPError{
		Code:     code,
		Message:  message,
		Internal: internal,
	}
}

type HTTPErrorInterface interface {
	Error() string
	GetCode() int
	SetCode(code int) error
	GetMessage() interface{}
	SetMessage(message interface{}) error
	GetInternal() error
	SetInternal(internal error) error
}

// HTTPError implements HTTP Error interface, default error object
type HTTPError struct {
	Code     int         `json:"code"`
	Message  interface{} `json:"message"`
	Internal error       `json:"-"` // Stores the error returned by an external dependency
}

// Error makes it compatible with `error` interface.
func (e *HTTPError) Error() string {
	if e.Internal == nil {
		return fmt.Sprintf("code=%d, message=%v", e.Code, e.Message)
	}
	return fmt.Sprintf("code=%d, message=%v, internal=%v", e.Code, e.Message, e.Internal)
}

func (e *HTTPError) GetCode() int {
	return e.Code
}

func (e *HTTPError) SetCode(code int) error {
	e.Code = code
	return nil
}

func (e *HTTPError) GetMessage() interface{} {
	return e.Message
}

func (e *HTTPError) SetMessage(message interface{}) error {
	e.Message = message
	return nil
}

func (e *HTTPError) GetInternal() error {
	return e.Internal
}

func (e *HTTPError) SetInternal(internal error) error {
	e.Internal = internal
	return nil
}

type ValidationResponse struct {
	Errors []*ValidationFieldError `json:"errors"`
}

type ValidationFieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

func CustomHTTPErrorHandler(err error, c echo.Context) {
	l := GetLogger(c)
	accept := GetAccept(c)

	l.Debug("CustomHTTPErrorHandler running", zap.Any("err", err), zap.String("accept", accept))

	code := 0
	if he, ok := err.(HTTPErrorInterface); ok {
		code = he.GetCode()
		if accept == "application/json" {
			c.JSON(code, he)
			return
		}
	}

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if accept == "application/json" {
			c.JSON(code, he)
			return
		}
	}

	if ve, ok := err.(validator.ValidationErrors); ok {
		validationError(ve, err, c)
		return
	}

	if code == 0 && err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		code = 404
	}

	switch code {
	case 401:
		unAuthorizedErrorHandler(err, c)
	case 403:
		forbiddenErrorHandler(err, c)
	case 404:
		notFoundErrorHandler(err, c)
	case 500:
		internalServerErrorHandler(err, c)
	default:
		l.Warn("customHTTPErrorHandler unknown error status code", zap.Error(err), zap.Int("statusCode", code), zap.String("path", c.Path()), zap.String("method", c.Request().Method), zap.Any("AuthenticatedUser", GetAuthenticatedUser(c)), zap.Any("roles", GetRoles(c)))
		c.JSON(http.StatusInternalServerError, &HTTPError{Code: 500, Message: "Unknown Error"})
	}
}

func forbiddenErrorHandler(err error, c echo.Context) error {
	accept := GetAccept(c)
	metadata := GetMetadata(c)
	l := GetLogger(c)

	l.Debug("forbiddenErrorHandler running", zap.Error(err), zap.String("accept", accept), zap.String("path", c.Path()), zap.String("method", c.Request().Method))

	switch accept {
	case "text/html":
		metadata.Set("title", "Forbidden")

		if err := c.Render(http.StatusForbidden, "403", &TemplateCTX{
			Ctx: c,
		}); err != nil {
			c.Logger().Error(err)
		}

		return nil
	default:
		c.JSON(http.StatusForbidden, err)
		return nil
	}
}

func unAuthorizedErrorHandler(err error, c echo.Context) error {
	l := GetLogger(c)
	accept := GetAccept(c)
	metadata := GetMetadata(c)

	l.Info("unAuthorizedErrorHandler running", zap.Error(err), zap.String("accept", accept), zap.String("path", c.Path()), zap.String("method", c.Request().Method), zap.Any("AuthenticatedUser", GetAuthenticatedUser(c)), zap.Any("roles", GetRoles(c)))

	switch accept {
	case "text/html":
		metadata.Set("title", "Forbidden")

		if err := c.Render(http.StatusUnauthorized, "401", &TemplateCTX{
			Ctx: c,
		}); err != nil {
			l.Error("unAuthorizedErrorHandler error rendering template", zap.Error(err), zap.String("accept", accept), zap.String("path", c.Path()), zap.String("method", c.Request().Method))
		}

		return nil
	default:
		c.JSON(http.StatusUnauthorized, err)
		return nil
	}
}

func notFoundErrorHandler(err error, c echo.Context) error {
	accept := GetAccept(c)
	metadata := GetMetadata(c)
	l := GetLogger(c)

	l.Debug("notFoundErrorHandler running", zap.Error(err), zap.String("accept", accept), zap.String("path", c.Path()), zap.String("method", c.Request().Method))

	switch accept {
	case "text/html":
		metadata.Set("title", "Not found")

		if err := c.Render(http.StatusNotFound, "404", &TemplateCTX{
			Ctx: c,
		}); err != nil {
			l.Error("notFoundErrorHandler error rendering template", zap.Error(err), zap.String("accept", accept), zap.String("path", c.Path()), zap.String("method", c.Request().Method))
		}
		return nil
	default:
		c.JSON(http.StatusNotFound, &HTTPError{Code: http.StatusNotFound, Message: "Not Found"})
		return nil
	}
}

func validationError(ve validator.ValidationErrors, err error, c echo.Context) error {
	accept := GetAccept(c)
	l := GetLogger(c)
	metadata := GetMetadata(c)

	l.Debug("validationError running", zap.Error(err), zap.String("accept", accept), zap.String("path", c.Path()), zap.String("method", c.Request().Method))

	resp := ValidationResponse{}

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var el ValidationFieldError
			el.Field = err.Field()
			el.Tag = err.Tag()
			el.Value = err.Param()
			el.Message = err.Error()
			resp.Errors = append(resp.Errors, &el)
		}
	}

	switch accept {
	case "text/html":
		metadata.Set("title", "Bad request")

		if err := c.Render(http.StatusInternalServerError, "400", &TemplateCTX{
			Ctx: c,
		}); err != nil {
			l.Error("validationError error rendering template", zap.Error(err), zap.String("accept", accept), zap.String("path", c.Path()), zap.String("method", c.Request().Method))
		}

		return nil
	default:
		return c.JSON(http.StatusBadRequest, resp)
	}
}

func internalServerErrorHandler(err error, c echo.Context) error {
	accept := GetAccept(c)
	l := GetLogger(c)
	metadata := GetMetadata(c)

	code := http.StatusInternalServerError
	if he, ok := err.(*HTTPError); ok {
		code = he.Code
	}

	l.Warn("internalServerErrorHandler error", zap.Error(err), zap.String("accept", accept), zap.String("path", c.Path()), zap.String("method", c.Request().Method), zap.Int("code", code))

	switch accept {
	case "text/html":
		metadata.Set("title", "Internal server error")

		if err := c.Render(code, "500", &TemplateCTX{
			Ctx: c,
		}); err != nil {
			l.Error("internalServerErrorHandler error rendering template", zap.Error(err), zap.String("accept", accept), zap.String("path", c.Path()), zap.String("method", c.Request().Method))
		}

		return nil
	default:
		if he, ok := err.(*HTTPError); ok {
			return c.JSON(he.Code, &HTTPError{Code: he.Code, Message: he.Message})
		}

		c.JSON(code, &HTTPError{Code: http.StatusInternalServerError, Message: "Internal Server Error"})
		return nil
	}
}
