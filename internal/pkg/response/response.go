package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	appErrors "github.com/rakib/job-portal-api/internal/pkg/errors"
)

type SuccessEnvelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponsePayload struct {
	Code    string                  `json:"code"`
	Message string                  `json:"message"`
	Details []appErrors.ErrorDetail `json:"details,omitempty"`
}

type ErrorEnvelope struct {
	Success bool                 `json:"success"`
	Error   ErrorResponsePayload `json:"error"`
}

func JSON(c *gin.Context, httpStatus int, message string, data interface{}) {
	c.JSON(httpStatus, SuccessEnvelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func OK(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, "OK", data)
}

func Created(c *gin.Context, data interface{}) {
	JSON(c, http.StatusCreated, "Created", data)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Error(c *gin.Context, err error) {
	if appErr, ok := err.(*appErrors.AppError); ok {
		c.JSON(appErr.HTTPStatus, ErrorEnvelope{
			Success: false,
			Error: ErrorResponsePayload{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			},
		})
		return
	}

	// Default internal server error
	c.JSON(http.StatusInternalServerError, ErrorEnvelope{
		Success: false,
		Error: ErrorResponsePayload{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		},
	})
}
