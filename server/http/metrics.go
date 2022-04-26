package http

import (
	"strconv"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/zhlls/go-common/metrics"
	"github.com/zhlls/go-common/utils"
)

func routerPathPrepare(path string, h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.SetUserValue("__router_path__", path)
		h(ctx)
	}
}

func metricsMiddleware(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Start timer
		start := time.Now()

		//spanCtx := GetTraceContext(ctx)
		//if spanCtx != context.Background() {
		//	span, nextCtx := opentracing.StartSpanFromContext(
		//		spanCtx, "metricsMiddleware")
		//	defer span.Finish()
		//	SetTraceContext(ctx, nextCtx)
		//}

		h(ctx)

		if utils.SliceToString(ctx.Path()) == metrics.DefaultURI {
			return
		}

		latency := time.Since(start)

		var fullPath string
		v := ctx.UserValue("__router_path__")
		if v != nil {
			fullPath = v.(string)
		}
		if fullPath == "" {
			fullPath = string(ctx.Path())
		}

		contentLength := ctx.Request.Header.ContentLength()
		if contentLength < 0 {
			contentLength = 0
		}
		metrics.CollectAPIRequestBytes(
			string(ctx.Method()),
			fullPath,
			strconv.Itoa(ctx.Response.StatusCode()),
			float64(contentLength),
		)
		metrics.CollectAPIRequestTotal(
			string(ctx.Method()),
			fullPath,
			strconv.Itoa(ctx.Response.StatusCode()),
		)
		metrics.CollectAPIResponseTime(
			string(ctx.Method()),
			fullPath,
			strconv.Itoa(ctx.Response.StatusCode()),
			latency.Seconds()*1000)
		metrics.CollectAPIResponseSize(
			string(ctx.Method()),
			fullPath,
			strconv.Itoa(ctx.Response.StatusCode()),
			float64(len(ctx.Response.Body())))
	}
}
