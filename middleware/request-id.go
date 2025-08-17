package middleware

import (
	gmw "github.com/Laisky/gin-middlewares/v6"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/helper"
)

func RequestId() func(c *gin.Context) {
	return func(c *gin.Context) {
		id := helper.GenRequestID()
		c.Set(helper.RequestIdKey, id)
		ctx := helper.SetRequestID(gmw.Ctx(c), id)
		c.Request = c.Request.WithContext(ctx)
		c.Header(helper.RequestIdKey, id)
		c.Next()
	}
}
