package auth

import (
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"
	kratos "github.com/ory/kratos-client-go"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
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

// RequireAuth checks if the request is authenticated with Ory Kratos, storing session information if so, and
// returning an error if not
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := c.(*config.Context)
		if cc.Session == nil || !*cc.Session.Active {
			return echo.NewHTTPError(http.StatusUnauthorized, "Not logged in")
		}
		return next(cc)
	}
}

// RequireRole checks if the authenticated user has the given role
func RequireRole(requiredRole string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := c.(*config.Context)
			if cc.Session == nil || !*cc.Session.Active {
				return echo.NewHTTPError(http.StatusUnauthorized, "Not logged in")
			}

			metadata, err := ParseMetadataPublic(cc.Session.Identity.MetadataPublic)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Error retrieving roles")
			}

			if !slices.Contains(metadata.Roles, requiredRole) {
				return echo.NewHTTPError(http.StatusForbidden, "Access denied")
			}

			return next(cc)
		}
	}
}
