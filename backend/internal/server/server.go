package server

import (
	"github.com/An-Owlbear/homecloud/backend/internal/api"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
)

func CreateServer() *echo.Echo {
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer docker.Close()

	e := echo.New()
	api.AddRoutes(e, docker)

	return e
}
