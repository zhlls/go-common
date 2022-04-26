package http

import (
	"fmt"

	"github.com/valyala/fasthttp"

	"github.com/zhlls/go-common/metrics"
	"github.com/zhlls/go-common/utils"
	"github.com/zhlls/go-common/version"
)

const (
	healthzPath = "/healthz"
	healthzLog  = "/healthz/log"
	healthzPing = "/healthz/ping"
	healthzInfo = "/info"
)

type routeInfo struct {
	path    string
	method  string
	handler fasthttp.RequestHandler
}

var routeList []routeInfo

func AddRouter(method, path string, h fasthttp.RequestHandler) {
	routeList = append(routeList, routeInfo{
		path:    path,
		method:  method,
		handler: h,
	})
}

func JsonNotFoundHandler(ctx *fasthttp.RequestCtx) {
	Failed(ctx, fasthttp.StatusNotFound, "not found this router")
}

func JsonMethodNotAllowedHandler(ctx *fasthttp.RequestCtx) {
	Failed(ctx, fasthttp.StatusMethodNotAllowed, "method not allowed")
}

func HandleHealthz(ctx *fasthttp.RequestCtx) {
	if ctx.IsHead() {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.Response.Header.Set("Content-Length", "0")
		return
	}

	OK(ctx, map[string]interface{}{
		"version":       version.Info.Version,
		"gitRevision":   version.Info.GitRevision,
		"user":          fmt.Sprintf("%s@%s", version.Info.User, version.Info.Host),
		"buildTime":     version.Info.BuildTime,
		"golangVersion": version.Info.GolangVersion,
		"buildStatus":   version.Info.BuildStatus,
	})
}

func shouldIgnore(ctx *fasthttp.RequestCtx) bool {
	uri := utils.SliceToString(ctx.Path())
	return uri == healthzInfo ||
		uri == metrics.DefaultURI ||
		uri == healthzPath ||
		uri == healthzLog ||
		uri == healthzPing
}
