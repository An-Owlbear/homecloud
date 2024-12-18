package api

import (
	"context"
	"fmt"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
)

type containerInfo struct {
	Id        string   `json:"id"`
	Name      []string `json:"name"`
	Container string   `json:"container"`
}

func AddRoutes(
	e *echo.Echo,
	docker *client.Client,
	queries *persistence.Queries,
	storeClient *apps.StoreClient,
) {
	e.GET("/", test(docker))
	e.GET("/db", db_test(queries))

	e.POST("/api/v1/packages/install", AddPackage(storeClient))
	e.POST("/api/v1/packages/update", CheckUpdates(storeClient))
}

func test(docker *client.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		containers, err := docker.ContainerList(context.Background(), container.ListOptions{
			All: true,
		})
		if err != nil {
			return c.String(500, err.Error())
		}

		var response []containerInfo
		for i := 0; i < len(containers); i++ {
			response = append(response, containerInfo{
				Id:        containers[i].ID,
				Name:      containers[i].Names,
				Container: containers[i].Image,
			})
		}
		return c.JSONPretty(200, response, "  ")
	}
}

func db_test(queries *persistence.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		response, err := queries.GetApps(context.Background())
		if err != nil {
			return c.String(500, err.Error())
		}

		fmt.Printf("%v+\n", response)
		return c.JSONPretty(200, response, "  ")
	}
}
