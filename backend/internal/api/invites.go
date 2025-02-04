package api

import (
	"encoding/json"
	"github.com/An-Owlbear/homecloud/backend/internal/auth"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/labstack/echo/v4"
	kratos "github.com/ory/kratos-client-go"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type invitationCodeRequest struct {
	UserId         string `json:"user_id"`
	InvitationCode string `json:"invitation_code"`
}

type webhookErrorContext struct {
	Value string `json:"value"`
	Any   string `json:"any"`
}

type webhookErrorMessageContents struct {
	Id          int                 `json:"id"`
	Text        string              `json:"text"`
	MessageType string              `json:"type"`
	Context     webhookErrorContext `json:"context"`
}

type webhookErrorMessage struct {
	InstancePtr string                        `json:"instance_ptr"`
	Messages    []webhookErrorMessageContents `json:"messages"`
}

type webhookError struct {
	Messages []webhookErrorMessage `json:"messages"`
}

func makeWebhookError(code int, message string) webhookError {
	return webhookError{
		Messages: []webhookErrorMessage{{
			Messages: []webhookErrorMessageContents{{
				Id:          code,
				Text:        message,
				MessageType: "error",
				Context:     webhookErrorContext{},
			}},
		}},
	}
}

// CheckInvitationCode checks if an invitation code is valid when registering
func CheckInvitationCode(queries *persistence.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Reads and parses request body
		reqBody, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, makeWebhookError(100, "Failed to read body"))
		}

		var invitationCode invitationCodeRequest
		if err = json.Unmarshal(reqBody, &invitationCode); err != nil {
			return c.JSON(http.StatusBadRequest, makeWebhookError(101, "Invalid JSON body"))
		}

		// Check code exists in the database
		code, err := queries.GetInviteCode(c.Request().Context(), invitationCode.InvitationCode)
		if err != nil {
			slog.Error(err.Error())
			return c.JSON(http.StatusBadRequest, makeWebhookError(102, "Invalid invitation code"))
		}

		if code.ExpiryDate.Before(time.Now()) {
			return c.JSON(http.StatusBadRequest, makeWebhookError(103, "Invite code expired"))
		}

		return c.JSON(http.StatusOK, webhookError{Messages: []webhookErrorMessage{}})
	}
}

// CompleteInvite removes used invite code
func CompleteInvite(queries *persistence.Queries, kratosAdmin kratos.IdentityAPI) echo.HandlerFunc {
	return func(c echo.Context) error {
		reqBody, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, makeWebhookError(100, "Failed to read body"))
		}

		var codeRequest invitationCodeRequest
		if err = json.Unmarshal(reqBody, &codeRequest); err != nil {
			return c.JSON(http.StatusBadRequest, makeWebhookError(101, "Invalid JSON body"))
		}

		inviteCode, err := queries.GetInviteCode(c.Request().Context(), codeRequest.InvitationCode)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, makeWebhookError(105, "Failed retrieving invite information"))
		}
		var rolesString []string
		err = json.Unmarshal([]byte(inviteCode.Roles), &rolesString)

		identityPatch := []kratos.JsonPatch{{
			Op:   "replace",
			Path: "/metadata_public",
			Value: auth.MetadataPublic{
				Roles: rolesString,
			},
		}}

		_, _, err = kratosAdmin.PatchIdentity(c.Request().Context(), codeRequest.UserId).
			JsonPatch(identityPatch).
			Execute()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, makeWebhookError(106, "Failed to patch invite"))
		}

		err = queries.RemoveInviteCode(c.Request().Context(), codeRequest.InvitationCode)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, makeWebhookError(104, "Failed to remove invite code"))
		}

		return c.JSON(http.StatusOK, webhookError{Messages: []webhookErrorMessage{}})
	}
}

type createInviteCodeRequest struct {
	ValidHours int `json:"valid_hours"`
}

func CreateInviteCode(queries *persistence.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		var request createInviteCodeRequest
		if err := c.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
		}

		result, err := queries.CreateInviteCode(c.Request().Context(), persistence.CreateInviteCodeParams{
			Hours:     1,
			Rolesjson: "[\"user\"]",
		})
		if err != nil {
			slog.Error(err.Error())
			return echo.NewHTTPError(http.StatusInternalServerError, "Error creating token")
		}

		return c.JSON(http.StatusOK, result)
	}
}
