package handler

import (
	"github.com/valyala/fasthttp"

	httpclient "github.com/zhlls/go-common/client/http"
	"github.com/zhlls/go-common/server/http"
)

func init() {
	http.AddRouter(fasthttp.MethodGet, "/demo", handleDemo)
}

func handleDemo(ctx *fasthttp.RequestCtx) {
	spanCtx := http.GetTraceContext(ctx)
	//span, nextCtx := opentracing.StartSpanFromContext(spanCtx, "handleDemo")
	//defer span.Finish()

	data, err := httpclient.SimpleTraceGet(spanCtx, "http://localhost:8081/info")
	if err != nil {
		http.Failed(ctx, 400, "info failed: "+err.Error())
		return
	}

	http.OK(ctx, "this is a demo! "+string(data))
}
