package main

import (
	syslog "log"
	"os"
	"os/signal"
	"syscall"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"

	_ "github.com/zhlls/go-common/examples/hello/handler"

	"github.com/zhlls/go-common/log"
	"github.com/zhlls/go-common/server/http"
	"github.com/zhlls/go-common/utils"
)

func main() {
	// recover from crash
	defer func() {
		if err := recover(); err != nil {
			stack := utils.Stack(3)
			syslog.Fatal(string(stack))
		}
	}()

	if err := initLog(); err != nil {
		syslog.Fatalf("Initialize log failed: \n%v", err)
	}

	// trace
	cfg, err := jaegerCfg.FromEnv()
	if err != nil {
		syslog.Fatalf("get trace config from env failed: \n%v", err)
	}
	cfg.Sampler.Type = jaeger.SamplerTypeConst
	cfg.Sampler.Param = 1
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	injector := jaegerCfg.Injector(opentracing.HTTPHeaders, zipkinPropagator)
	extractor := jaegerCfg.Extractor(opentracing.HTTPHeaders, zipkinPropagator)
	// Zipkin shares span ID between client and server spans; it must be enabled via the following option.
	//zipkinSharedRPCSpan := jaegerCfg.ZipkinSharedRPCSpan(true)
	closeTrace, err := cfg.InitGlobalTracer("hello",
		injector, extractor,
		//zipkinSharedRPCSpan,
	)
	if err != nil {
		syslog.Fatalf("Init trace failed: \n%v", err)
	}
	defer closeTrace.Close()

	// common metrics
	//go metrics.Init("hello")

	// start http server
	httpAddr := ":8082"
	httpServer := http.NewServer(httpAddr)
	httpServer.SetPathPrefix("")
	httpServer.NotFound = http.JsonNotFoundHandler
	go func() {
		err := httpServer.Start()
		if err != nil {
			os.Exit(-1)
		}
	}()

	waitStop()
}

func waitStop() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	sig := <-sc
	log.Infof("exit: signal=<%d>.", sig)

	switch sig {
	case syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT:
		log.Info("exit: bye :-).")
		os.Exit(0)
	default:
		log.Info("exit: bye :-(.")
		os.Exit(1)
	}
}

func initLog() error {

	loggingOptions := log.DefaultOptions()
	loggingOptions.Format = log.ConsoleFormat
	loggingOptions.ServiceName = "hello"

	level := log.InfoLevel

	loggingOptions.SetOutputLevel(log.DefaultScopeName, level)
	loggingOptions.SetLogCallers(log.DefaultScopeName, true)
	loggingOptions.SetStackTraceLevel(log.DefaultScopeName, log.ErrorLevel)
	if err := log.Configure(loggingOptions); err != nil {
		return err
	}

	return nil
}
