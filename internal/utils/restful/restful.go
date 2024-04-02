package restful

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type apiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

func ParseOptions(options ...interface{}) (string, map[string]interface{}) {
	msg := ""
	data := map[string]interface{}{}

	for _, value := range options {
		switch v := value.(type) {
		case map[string]interface{}:
			data = v
		case gin.H:
			data = v
		case string:
			msg = v
		}
	}
	return msg, data
}

func Ok(c *gin.Context, options ...interface{}) {
	msg, data := ParseOptions(options...)
	c.JSON(http.StatusOK, apiResponse{
		Code:    200,
		Message: msg,
		Data:    data,
	})
}

func ParamErr(c *gin.Context, options ...interface{}) {
	msg, data := ParseOptions(options...)
	c.JSON(http.StatusOK, apiResponse{
		Code:    400,
		Message: msg,
		Data:    data,
	})
}

func UnLoginErr(c *gin.Context, options ...interface{}) {
	msg, data := ParseOptions(options...)

	c.JSON(http.StatusOK, apiResponse{
		Code:    401,
		Message: msg,
		Data:    data,
	})
}
