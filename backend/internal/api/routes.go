package api

import (
	"github.com/An-Owlbear/homecloud/backend/internal/auth"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	hydra "github.com/ory/hydra-client-go/v2"
	kratos "github.com/ory/kratos-client-go"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
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
	appDataHandler *persistence.AppDataHandler,
	serverConfig config.Config,
) {
	apiNoAuth := e.Group("/api")
	api := apiNoAuth.Group("")
	api.Use(auth.RequireAuth)
	apiAdmin := api.Group("", auth.RequireRole("admin"))

	apiNoAuth.GET("/v1/check", PortForwardTest())

	api.GET("/v1/packages", ListPackages(storeClient))
	api.GET("/v1/packages/:id", GetPackage(storeClient))
	api.GET("/v1/packages/search", SearchPackages(storeClient))
	apiAdmin.POST("/v1/packages/:appId/install", AddPackage(storeClient, queries, docker, hydraAdmin, serverConfig.Host, appDataHandler))
	apiAdmin.POST("/v1/packages/update", CheckUpdates(storeClient))
	api.GET("/v1/packages/categories", ListCategories(storeClient))

	api.GET("/v1/apps", ListApps(queries))
	apiAdmin.POST("/v1/apps/:appId/start", StartApp(docker, queries, hosts))
	apiAdmin.POST("/v1/apps/:appId/stop", StopApp(docker))
	apiAdmin.POST("/v1/apps/:appId/uninstall", UninstallApp(queries, docker))
	apiAdmin.POST("/v1/apps/update", UpdateApps(docker, storeClient, queries, serverConfig.Host))

	apiNoAuth.POST("/v1/invites/check", CheckInvitationCode(queries))
	apiAdmin.PUT("/v1/invites", CreateInviteCode(queries))
	apiNoAuth.POST("/v1/invites/complete", CompleteInvite(queries, kratosIdentityAPI))

	apiAdmin.GET("/v1/users", ListUsers(kratosIdentityAPI))
	apiAdmin.DELETE("/v1/users/:id", DeleteUser(kratosIdentityAPI))
	apiAdmin.POST("/v1/users/:id/reset_password", ResetPassword(kratosIdentityAPI))

	e.GET("/auth/login", Login(kratosClient, serverConfig.Ory))
	e.GET("/auth/registration", Registration(kratosClient, serverConfig.Ory))
	e.GET("/auth/settings", Settings(kratosClient, serverConfig.Ory))
	e.GET("/auth/recovery", Recovery(kratosClient))
	e.GET("/auth/oidc", OidcConsent(hydraAdmin))
	e.Static("/assets", "assets")
	e.GET("/assets/data/*", staticFilter(serverConfig.Storage.DataPath, "^db\\/data\\/.+\\/icon.png$"))
}

func PortForwardTest() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.NoContent(204)
	}
}
