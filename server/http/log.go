package http

import (
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"github.com/zhlls/go-common/log"
)

func logMiddleware(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		var start time.Time
		var raw []byte
		if log.DebugEnabled() {
			// Start timer
			start = time.Now()
			//path := ctx.Path()
			raw = ctx.Request.RequestURI()
		}

		//spanCtx := GetTraceContext(ctx)
		//if spanCtx != context.Background() {
		//	span, nextCtx := opentracing.StartSpanFromContext(
		//		spanCtx, "logMiddleware")
		//	defer span.Finish()
		//	SetTraceContext(ctx, nextCtx)
		//}

		h(ctx)

		if log.DebugEnabled() {
			if shouldIgnore(ctx) {
				return
			}

			end := time.Now()
			latency := end.Sub(start)
			clientIP := ctx.RemoteIP().String()
			method := ctx.Method()
			statusCode := ctx.Response.StatusCode()
			log.Debug("http request",
				zap.Time("end", end),
				zap.Int("status", statusCode),
				zap.Duration("latency", latency),
				zap.String("client", clientIP),
				zap.ByteString("method", method),
				zap.ByteString("path", raw),
				zap.Int("response-size", len(ctx.Response.Body())),
			)
		}
	}
}
