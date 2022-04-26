package http

import (
	"time"

	"github.com/valyala/fasthttp"

	"github.com/zhlls/go-common/log"
	"github.com/zhlls/go-common/utils"
)

type PanicHandler func(*fasthttp.RequestCtx, interface{})

func recoveryHandler(handler ...PanicHandler) PanicHandler {
	return func(ctx *fasthttp.RequestCtx, info interface{}) {
		stack := utils.Stack(3)
		httprequest := ctx.Request.String()
		log.Errorf(
			"\n\n\x1b[31m[Recovery] %s panic recovered:\n%s\n%s\n%s\033[0m",
			time.Now().Format("2006/01/02 - 15:04:05.999"),
			httprequest,
			info,
			stack,
		)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)

		for _, h := range handler {
			if h == nil {
				continue
			}
			h(ctx, info)
		}
	}
}
