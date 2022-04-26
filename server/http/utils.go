package http

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"github.com/zhlls/go-common/log"
)

var (
	StrContentEncoding = []byte("Content-Encoding")
	StrGzip            = []byte("gzip")

	//strContentType = []byte("Content-Type")
	StrApplicationJSON = []byte("application/json")

	errPathArgRequired = errors.New("path arg required")
	errPathArgInvalid  = errors.New("path arg invalid")
)

func GetPathArg(ctx *fasthttp.RequestCtx, name string) (string, error) {
	arg := ctx.UserValue(name)
	if arg == nil {
		log.Error(errPathArgRequired.Error(),
			zap.String("name", name))
		return "", errPathArgRequired
	}
	value, ok := arg.(string)
	if !ok || value == "" {
		log.Error(errPathArgInvalid.Error(),
			zap.String("name", name),
			zap.Any("value", arg))
		return "", errPathArgInvalid
	}

	return value, nil
}

type errorMsg struct {
	Msg string `json:"msg"`
}
type failedMsg struct {
	Msg    string      `json:"msg"`
	Result interface{} `json:"result"`
}
type resultMap map[string]interface{}
type resultArray []string
type resultString string
type resultInt int

func BadRequestMap(ctx *fasthttp.RequestCtx, err error) {
	doJSONWrite(ctx, fasthttp.StatusBadRequest, failedMsg{
		Msg:    err.Error(),
		Result: resultMap{},
	})
}
func BadRequestArray(ctx *fasthttp.RequestCtx, err error) {
	doJSONWrite(ctx, fasthttp.StatusBadRequest, failedMsg{
		Msg:    err.Error(),
		Result: resultArray{},
	})
}

func BadRequest(ctx *fasthttp.RequestCtx, data interface{}) {
	doJSONWrite(ctx, fasthttp.StatusBadRequest, data)
}

func InternalServerError(ctx *fasthttp.RequestCtx, data interface{}) {
	doJSONWrite(ctx, fasthttp.StatusInternalServerError, data)
}

func OK(ctx *fasthttp.RequestCtx, data interface{}) {
	doJSONWrite(ctx, fasthttp.StatusOK, data)
}

func DeleteOk(ctx *fasthttp.RequestCtx) {
	doJSONWrite(ctx, fasthttp.StatusNoContent, nil)
}

func CreateOk(ctx *fasthttp.RequestCtx, data interface{}) {
	doJSONWrite(ctx, fasthttp.StatusCreated, data)
}

func DiyOk(ctx *fasthttp.RequestCtx, statusCode int, data interface{}) {
	doJSONWrite(ctx, statusCode, data)
}

func Unauthorized(ctx *fasthttp.RequestCtx, data interface{}) {
	doJSONWrite(ctx, fasthttp.StatusUnauthorized, data)
}

func Forbidden(ctx *fasthttp.RequestCtx, data interface{}) {
	doJSONWrite(ctx, fasthttp.StatusForbidden, data)
}

func Failed(ctx *fasthttp.RequestCtx, code int, msg string) {
	//ctx.WriteString("{\"msg\":\""+msg+"\"}")
	doJSONWrite(ctx, code, errorMsg{msg})
}
func Success(ctx *fasthttp.RequestCtx, obj interface{}) {
	doJSONWrite(ctx, fasthttp.StatusOK, obj)
}

func doJSONWrite(ctx *fasthttp.RequestCtx, code int, obj interface{}) {
	ctx.SetContentTypeBytes(StrApplicationJSON)
	ctx.SetStatusCode(code)
	start := time.Now()
	if err := json.NewEncoder(ctx).Encode(obj); err != nil {
		elapsed := time.Since(start)
		log.Error("json encode error",
			zap.Duration("elapsed", elapsed),
			zap.Error(err))
		log.Debug("json encode error",
			zap.Any("obj", obj))
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
}
