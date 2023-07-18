package bolo

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type Action func(ctx echo.Context) (Response, error)

type Response interface {
	GetData() any
	GetStatusCode() int
}

type DefaultResponse struct {
	Data   any
	Status int
}

func (r *DefaultResponse) GetData() any {
	return r.Data
}

func (r *DefaultResponse) GetStatusCode() int {
	if r.Status == 0 {
		return http.StatusOK
	}

	return r.Status
}

type Route struct {
	Method     string
	Path       string
	Action     Action
	Permission string
	AcceptOnly string
	Template   string
	Layout     string
	Model      interface{}
}

type responseFormatter func(app App, c echo.Context, r *Route, resp Response) error

// NegotiateContentType returns the best offered content type for the request's
// Accept header. If two offers match with equal weight, then the more specific
// offer is preferred.  For example, text/* trumps */*. If two offers match
// with equal weight and specificity, then the offer earlier in the list is
// preferred. If no offers match, then defaultOffer is returned.
func NegotiateContentType(r *http.Request, offers []string, defaultOffer string) string {
	bestOffer := defaultOffer
	bestQ := -1.0
	bestWild := 3
	specs := ParseAccept(r.Header, "Accept")
	for _, offer := range offers {
		for _, spec := range specs {
			switch {
			case spec.Q == 0.0:
				// ignore
			case spec.Q < bestQ:
				// better match found
			case spec.Value == "*/*":
				if spec.Q > bestQ || bestWild > 2 {
					bestQ = spec.Q
					bestWild = 2
					bestOffer = offer
				}
			case strings.HasSuffix(spec.Value, "/*"):
				if strings.HasPrefix(offer, spec.Value[:len(spec.Value)-1]) &&
					(spec.Q > bestQ || bestWild > 1) {
					bestQ = spec.Q
					bestWild = 1
					bestOffer = offer
				}
			default:
				if spec.Value == offer &&
					(spec.Q > bestQ || bestWild > 0) {
					bestQ = spec.Q
					bestWild = 0
					bestOffer = offer
				}
			}
		}
	}
	return bestOffer
}

func IsPublicRoute(url string) bool {
	return strings.HasPrefix(url, "/health") || strings.HasPrefix(url, "/public")
}

func HealthCheckHandler(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

type CustomBinder struct{}

func (cb *CustomBinder) Bind(i interface{}, c echo.Context) (err error) {
	// You may use default binder
	db := &echo.DefaultBinder{}
	if err = db.Bind(i, c); err != echo.ErrUnsupportedMediaType {
		return
	}

	return
}
