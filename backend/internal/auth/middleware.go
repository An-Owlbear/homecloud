package auth

import (
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/labstack/echo/v4"
	kratos "github.com/ory/kratos-client-go"
	"net/http"
)

// Middleware sets the session in the context
func Middleware(kratosFrontend kratos.FrontendAPI) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := c.(*config.Context)

			session, _, err := kratosFrontend.
				ToSession(cc.Request().Context()).
				Cookie(cc.Request().Header.Get("Cookie")).
				Execute()
			if err != nil {
				return next(cc)
			}
			cc.Session = session
			return next(cc)
		}
	}
}

func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := c.(*config.Context)
		if cc.Session == nil || !*cc.Session.Active {
			return echo.NewHTTPError(http.StatusUnauthorized, "Not logged in")
		}
		return next(cc)
	}
}
