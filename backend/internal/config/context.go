package config

import (
	"github.com/labstack/echo/v4"
	kratos "github.com/ory/kratos-client-go"
)

type Context struct {
	echo.Context
	Session *kratos.Session
}

func ContextMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := &Context{Context: c}
		return next(cc)
	}
}
