package metrics

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

var (
	httpRequestsTotal   *Metric
	httpRequestDuration *Metric
	httpRequestsInFlight *Metric
)

func init() {
	httpRequestsTotal = DefaultRegistry.Register(
		"http_requests_total",
		TypeCounter,
		"Total number of HTTP requests",
	)

	httpRequestDuration = DefaultRegistry.Register(
		"http_request_duration_seconds",
		TypeHistogram,
		"HTTP request duration in seconds",
	)

	httpRequestsInFlight = DefaultRegistry.Register(
		"http_requests_in_flight",
		TypeGauge,
		"Number of HTTP requests currently being processed",
	)
}

func Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		method := c.Method()
		path := c.Path()

		labels := Labels{
			"method": method,
			"path":   path,
		}

		httpRequestsInFlight.Set(httpRequestsInFlightCount()+1, labels)
		defer func() {
			httpRequestsInFlight.Set(httpRequestsInFlightCount()-1, labels)
		}()

		err := c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())

		labels["status"] = status

		httpRequestsTotal.Inc(labels)
		httpRequestDuration.Observe(duration, labels)

		return err
	}
}

func httpRequestsInFlightCount() float64 {
	seriesList := httpRequestsInFlight.ListSeries()
	if len(seriesList) == 0 {
		return 0
	}
	return 0
}

func Inc(name string, labels Labels) {
	if m, ok := DefaultRegistry.Get(name); ok {
		m.Inc(labels)
	}
}

func Add(name string, v float64, labels Labels) {
	if m, ok := DefaultRegistry.Get(name); ok {
		m.Add(v, labels)
	}
}

func Set(name string, v float64, labels Labels) {
	if m, ok := DefaultRegistry.Get(name); ok {
		m.Set(v, labels)
	}
}

func Observe(name string, v float64, labels Labels) {
	if m, ok := DefaultRegistry.Get(name); ok {
		m.Observe(v, labels)
	}
}

func RegisterCounter(name, help string) *Metric {
	return DefaultRegistry.Register(name, TypeCounter, help)
}

func RegisterGauge(name, help string) *Metric {
	return DefaultRegistry.Register(name, TypeGauge, help)
}

func RegisterHistogram(name, help string) *Metric {
	return DefaultRegistry.Register(name, TypeHistogram, help)
}
