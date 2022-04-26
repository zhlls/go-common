package http

import (
	"net"
	"time"

	fasthttprouter "github.com/fasthttp/router"
	sentryfasthttp "github.com/getsentry/sentry-go/fasthttp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"github.com/valyala/fasthttp/pprofhandler"
	"go.uber.org/zap"

	"github.com/zhlls/go-common/log"
	"github.com/zhlls/go-common/metrics"
)

const defaultMaxRequestBodySize = 64 * 1024 * 1024

type Server struct {
	addr string
	http *fasthttp.Server

	pathPrefix string

	NotFound         fasthttp.RequestHandler
	MethodNotAllowed fasthttp.RequestHandler

	middleware   []Middleware
	disableTrace bool
	enableSentry bool

	panicHandlers []PanicHandler
}

//func timeoutMiddleware(h fasthttp.RequestHandler) fasthttp.RequestHandler {
//	return fasthttp.TimeoutHandler(h, time.Minute, "request timeout")
//}

func (s *Server) SetPathPrefix(prefix string) {
	s.pathPrefix = prefix
}

func (s *Server) DisableTrace(disable bool) {
	s.disableTrace = disable
}

func (s *Server) SetSentry(status bool) {
	s.enableSentry = status
}

func (s *Server) WithPanicHandlers(handler ...PanicHandler) *Server {
	s.panicHandlers = handler
	return s
}

func (s *Server) Start() error {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("http server crash, errors: \n %+v", err)
		}
	}()

	// router
	router := fasthttprouter.New()
	router.PanicHandler = recoveryHandler(s.panicHandlers...)
	router.NotFound = s.NotFound
	router.MethodNotAllowed = s.MethodNotAllowed

	router.GET(metrics.DefaultURI, fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler()))
	router.GET("/debug/pprof/{name:*}", pprofhandler.PprofHandler)

	router.HEAD(healthzPath, HandleHealthz)
	router.HEAD(healthzLog, HandleHealthz)
	router.HEAD(healthzPing, HandleHealthz)
	router.GET(healthzPath, HandleHealthz)
	router.GET(healthzLog, HandleHealthz)
	router.GET(healthzPing, HandleHealthz)
	router.GET(healthzInfo, HandleHealthz)

	for _, ri := range routeList {
		fullPath := s.pathPrefix + ri.path
		handle := routerPathPrepare(fullPath, ri.handler)
		if s.enableSentry {
			handle = s.EnableSentry(handle)
		}
		router.Handle(ri.method, fullPath, handle)
	}

	s.http.Handler = s.finallyHandler(router)

	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Error("http: listen addr failed",
			zap.String("addr", s.addr),
			zap.Error(err))
		return err
	}
	log.Info("http server listening at " + s.addr)

	if err := s.http.Serve(lis); err != nil {
		log.Error("Error in http Serve", zap.Error(err))
		return err
	}

	return nil
}

func (s *Server) Stop() {
	err := s.http.Shutdown()
	if err != nil {
		log.Warn("shutdown http error", zap.Error(err))
	}
}

type Middleware func(fasthttp.RequestHandler) fasthttp.RequestHandler

func (s *Server) Use(m ...Middleware) {
	s.middleware = append(s.middleware, m...)
}

func (s *Server) EnableSentry(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	// Later in the code
	sentryHandler := sentryfasthttp.New(sentryfasthttp.Options{
		// Repanic: false,
		// WaitForDelivery: true,
	})
	return sentryHandler.Handle(h)
}

func (s *Server) finallyHandler(router *fasthttprouter.Router) fasthttp.RequestHandler {
	rh := router.Handler
	for i := len(s.middleware) - 1; i >= 0; i-- {
		rh = s.middleware[i](rh)
	}
	return rh
}

func NewServer(addr string) *Server {
	s := &Server{
		addr:       addr,
		pathPrefix: "/rest",
		middleware: []Middleware{
			metricsMiddleware,
			logMiddleware,
			// timeoutMiddleware,
			fasthttp.CompressHandler,
		},
	}

	if !s.disableTrace {
		s.middleware = append([]Middleware{tracerMiddleware}, s.middleware...)
	}

	s.http = &fasthttp.Server{
		ReadTimeout:        120 * time.Second,
		MaxRequestBodySize: defaultMaxRequestBodySize,
	}

	return s
}
