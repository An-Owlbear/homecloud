package api

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"io"
	"log/slog"
	"net/http"
)

type invitationCodeRequest struct {
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

func CheckInvitationCode() echo.HandlerFunc {
	return func(c echo.Context) error {
		reqBody, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, makeWebhookError(100, "Failed to read body"))
		}
		slog.Info(string(reqBody))

		var invitationCode invitationCodeRequest
		if err = json.Unmarshal(reqBody, &invitationCode); err != nil {
			return c.JSON(http.StatusBadRequest, makeWebhookError(101, "Invalid JSON body"))
		}

		slog.Info(invitationCode.InvitationCode)
		if invitationCode.InvitationCode == "PLACEHOLDER" {
			return c.JSON(http.StatusOK, webhookError{Messages: []webhookErrorMessage{}})
		}

		return c.JSON(http.StatusBadRequest, makeWebhookError(102, "Invalid invitation code"))
	}
}
