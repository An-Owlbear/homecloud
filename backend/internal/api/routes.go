package api

import (
	"context"
	"fmt"
	"github.com/An-Owlbear/homecloud/backend/internal/auth"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	hydra "github.com/ory/hydra-client-go/v2"
	kratos "github.com/ory/kratos-client-go"

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
	hosts *apps.Hosts,
	hydraAdmin *hydra.APIClient,
	kratosClient *kratos.APIClient,
	kratosIdentityAPI kratos.IdentityAPI,
	hostConfig config.Host,
) {
	apiNoAuth := e.Group("/api")
	api := apiNoAuth.Group("")
	api.Use(auth.RequireAuth)
	apiAdmin := api.Group("", auth.RequireRole("admin"))

	api.GET("/", test(docker))
	api.GET("/db", db_test(queries))

	apiAdmin.POST("/v1/packages/:appId/install", AddPackage(storeClient, queries, docker, hydraAdmin, hostConfig))
	apiAdmin.POST("/v1/packages/update", CheckUpdates(storeClient))

	api.GET("/v1/apps", ListApps(queries))
	apiAdmin.POST("/v1/apps/:appId/start", StartApp(docker, queries, hosts))
	apiAdmin.POST("/v1/apps/:appId/stop", StopApp(docker))
	apiAdmin.POST("/v1/apps/:appId/uninstall", UninstallApp(queries, docker))
	apiAdmin.POST("/v1/apps/update", UpdateApps(docker, storeClient, queries))

	apiNoAuth.POST("/v1/invites/check", CheckInvitationCode(queries))
	apiAdmin.PUT("/v1/invites", CreateInviteCode(queries))
	apiNoAuth.POST("/v1/invites/complete", CompleteInvite(queries, kratosIdentityAPI))

	apiAdmin.GET("/v1/users", ListUsers(kratosIdentityAPI))
	apiAdmin.DELETE("/v1/users/:id", DeleteUser(kratosIdentityAPI))
	apiAdmin.POST("/v1/users/:id/reset_password", ResetPassword(kratosIdentityAPI))

	e.GET("/auth/login", Login(kratosClient))
	e.GET("/auth/registration", Registration(kratosClient))
	e.GET("/auth/settings", Settings(kratosClient))
	e.GET("/auth/recovery", Recovery(kratosClient))
	e.GET("/auth/oidc", OidcConsent(hydraAdmin))
	e.Static("/assets", "assets")
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
