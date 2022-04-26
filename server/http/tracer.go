package http

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"github.com/zhlls/go-common/log"
)

const spanCtxKey = "__span_context__"

type HeadersCarrierWriter fasthttp.ResponseHeader
type HeadersCarrierReader fasthttp.RequestHeader

// Set conforms to the TextMapWriter interface.
func (c *HeadersCarrierWriter) Set(key, val string) {
	h := (*fasthttp.ResponseHeader)(c)
	h.Set(key, val)
}

// ForeachKey conforms to the TextMapReader interface.
func (c *HeadersCarrierReader) ForeachKey(handler func(key, val string) error) error {
	h := (*fasthttp.RequestHeader)(c)
	h.VisitAll(func(key, value []byte) {
		if err := handler(string(key), string(value)); err != nil {
			return
		}
	})
	return nil
}

func tracerMiddleware(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if shouldIgnore(ctx) {
			h(ctx)
			return
		}

		spanCtx, err := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			(*HeadersCarrierReader)(&ctx.Request.Header))
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			// Optionally record something about err here
			log.Debug("trace extract failed",
				zap.ByteString("method", ctx.Method()),
				zap.ByteString("path", ctx.Request.RequestURI()),
				zap.Error(err),
			)
		}
		// just http, opentracing.ChildOf is enough, no need ext.RPCServerOption
		sp := opentracing.GlobalTracer().StartSpan(
			"tracerMiddleware", opentracing.ChildOf(spanCtx))
		ext.HTTPMethod.Set(sp, string(ctx.Method()))
		ext.HTTPUrl.Set(sp, ctx.URI().String())
		ext.Component.Set(sp, "fasthttp")

		nextCtx := opentracing.ContextWithSpan(ctx, sp)
		SetTraceContext(ctx, nextCtx)

		h(ctx)

		ext.HTTPStatusCode.Set(sp, uint16(ctx.Response.StatusCode()))
		sp.Finish()
	}
}

func SetTraceContext(ctx *fasthttp.RequestCtx, spanCtx context.Context) {
	ctx.SetUserValue(spanCtxKey, spanCtx)
}

func GetTraceContext(ctx *fasthttp.RequestCtx) context.Context {
	v := ctx.UserValue(spanCtxKey)
	spanCtx, ok := v.(context.Context)
	if ok {
		return spanCtx
	} else {
		return context.Background()
	}
}
