package http

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/opentracing/opentracing-go"
)

var client = http.Client{
	Transport: &transport{
		http.DefaultTransport,
	},
}

func SimpleTraceGet(ctx context.Context, uri string) ([]byte, error) {
	return SimpleTraceDo(ctx, http.MethodGet, uri, nil)
}

func SimpleTracePost(ctx context.Context, uri string, body io.Reader) ([]byte, error) {
	return SimpleTraceDo(ctx, http.MethodPost, uri, body)
}

func SimpleTraceDo(ctx context.Context, method, uri string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, uri, body)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

type transport struct {
	http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	span, nextCtx := opentracing.StartSpanFromContext(
		req.Context(), "http.client")
	defer span.Finish()

	err := opentracing.GlobalTracer().Inject(
		opentracing.SpanFromContext(nextCtx).Context(), opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header))
	if err != nil {}

	return t.RoundTripper.RoundTrip(req)
}
