package api

import (
	"strings"

	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	hydra "github.com/ory/hydra-client-go/v2"
	kratos "github.com/ory/kratos-client-go"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/auth"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/An-Owlbear/homecloud/backend/internal/storage"
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
	appDataHandler *storage.AppDataHandler,
	serverConfig config.Config,
	launcherProxy echo.MiddlewareFunc,
) {
	apiNoAuth := e.Group("/api")
	api := apiNoAuth.Group("")
	api.Use(auth.RequireAuth)
	apiAdmin := api.Group("", auth.RequireRole("admin"))

	apiNoAuth.GET("/v1/check", PortForwardTest())

	api.GET("/v1/packages", ListPackages(queries))
	api.GET("/v1/packages/:id", GetPackage(queries))
	api.GET("/v1/packages/search", SearchPackages(queries))
	apiAdmin.POST("/v1/packages/:appId/install", AddPackage(storeClient, queries, docker, hydraAdmin, serverConfig.Ory, serverConfig.Host, serverConfig.Storage, appDataHandler))
	apiAdmin.POST("/v1/packages/update", CheckUpdates(storeClient, queries))
	api.GET("/v1/packages/categories", ListCategories(queries))

	api.GET("/v1/apps", ListApps(queries))
	apiAdmin.POST("/v1/apps/:appId/start", StartApp(docker, queries, hosts, appDataHandler, serverConfig.Host, serverConfig.Ory))
	apiAdmin.POST("/v1/apps/:appId/stop", StopApp(docker, queries))
	apiAdmin.POST("/v1/apps/:appId/uninstall", UninstallApp(queries, docker, hydraAdmin))
	apiAdmin.POST("/v1/apps/update", UpdateApps(docker, storeClient, queries, serverConfig.Ory, serverConfig.Host, serverConfig.Storage))
	apiAdmin.POST("/v1/apps/:appId/backup", BackupApp(docker, serverConfig.Storage))

	apiNoAuth.POST("/v1/invites/check", CheckInvitationCode(queries))
	apiAdmin.POST("/v1/invites", CreateInviteCode(queries))
	apiNoAuth.POST("/v1/invites/complete", CompleteInvite(queries, kratosIdentityAPI))

	apiAdmin.GET("/v1/users", ListUsers(kratosIdentityAPI))
	apiAdmin.DELETE("/v1/users/:id", DeleteUser(kratosIdentityAPI))
	apiAdmin.POST("/v1/users/:id/reset_password", ResetPassword(kratosIdentityAPI))

	apiAdmin.GET("/v1/backup/devices", ListExternalStorage())

	e.GET("/auth/login", Login(kratosClient, serverConfig.Ory))
	e.GET("/auth/registration", Registration(kratosClient, serverConfig.Ory))
	e.GET("/auth/settings", Settings(kratosClient, serverConfig.Ory))
	e.GET("/auth/recovery", Recovery(kratosClient))
	e.GET("/auth/oidc", OidcConsent(hydraAdmin))
	e.GET("/auth/setup", InitialSetup(kratosIdentityAPI, queries))
	e.Static("/assets", "assets")
	e.GET("/assets/data/*", staticFilter(serverConfig.Storage.DataPath, "^db\\/data\\/.+\\/icon.png$"))

	// Proxies update urls to launcher on host system
	launcherApi := apiAdmin.Group("/v1/update")
	launcherApi.Use(launcherProxy)

	// Serves static SPA files
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:  "spa",
		HTML5: true,
		Skipper: func(c echo.Context) bool {
			for _, prefix := range []string{"/api", "/auth", "/assets"} {
				if strings.HasPrefix(c.Path(), prefix) {
					return true
				}
			}
			return false
		},
	}))
}

func PortForwardTest() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.NoContent(204)
	}
}
