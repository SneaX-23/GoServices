package telemetry

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Wraps an http.Handler with opentelemetry instrumentation
func HTTPMiddleware(handler http.Handler) http.Handler {
	return otelhttp.NewHandler(handler, "http-server",
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
	)
}

// Wraps an http.HandlerFunc with OpenTelemetry instrumentation
func WrapHandlerFunc(pattern string, handler http.HandlerFunc) (string, http.Handler) {
	return pattern, otelhttp.NewHandler(handler, pattern)
}
