package utils

import (
	"log/slog"
	"net/http"

	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
)

//  Centralised error response

// HandleError is the single function all handlers call when a service or decode
// step returns an error.
//
// Flow:
//   - If err is *apperror.AppError → use its Status, Code, Message
//   - 5xx errors → logged at Error level with the underlying cause
//   - 4xx errors → logged at Warn level (client mistake, not server fault)
//   - Unknown errors → always 500, never leak internals to the client
// 

func HandleError(w http.ResponseWriter, logger *slog.Logger, err error) {

	var appErr *apperror.AppError

	if apperror.As(err, &appErr) {

		// SAFE LOGGER USAGE
		if logger != nil {

			if appErr.Status >= 500 {
				logger.Error(appErr.Message,
					"code", appErr.Code,
					"cause", appErr.Err,
				)
			} else {
				logger.Warn(appErr.Message,
					"code", appErr.Code,
				)
			}
		}

		WriteJSON(w, appErr.Status, Envelope{
			"error":   appErr.Code,
			"message": appErr.Message,
		})
		return
	}

	// Unknown error — SAFE LOGGER
	if logger != nil {
		logger.Error("unhandled error", "cause", err)
	}

	WriteJSON(w, http.StatusInternalServerError, Envelope{
		"error":   "internal_error",
		"message": "an unexpected error occurred",
	})
}

// ServerError kept for backward compatibility — prefer HandleError.
func ServerError(w http.ResponseWriter, logger *slog.Logger, msg string, err error) {
	HandleError(w, logger, apperror.Internal(msg, err))
}

// BadRequest kept for backward compatibility — prefer HandleError.
func BadRequest(w http.ResponseWriter, logger *slog.Logger, msg string, _ error) {
	HandleError(w, logger, apperror.BadRequest("bad_request", msg))
}
