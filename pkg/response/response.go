package response

import (
	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func JSONError(c *gin.Context, status int, err string, msg ...string) {
	resp := ErrorResponse{Error: err}
	if len(msg) > 0 {
		resp.Message = msg[0]
	}
	c.JSON(status, resp)
}

func JSONSuccess(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}
