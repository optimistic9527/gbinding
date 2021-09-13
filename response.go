package gbinding

import (
	"log"

	"github.com/gin-gonic/gin"
)

type ResponseHandler func(ctx *gin.Context, data interface{}, err error)

var responseFn ResponseHandler

//SetGlobalResponse Customizing the global return
func SetGlobalResponse(rsFunc ResponseHandler) {
	if rsFunc == nil {
		log.Panic("set global response can't null")
	}
	responseFn = rsFunc
}

type response struct {
	hasData bool
}

func (r *response) Return(ctx *gin.Context, data interface{}, err error) {
	if responseFn != nil {
		responseFn(ctx, data, err)
	}
}
