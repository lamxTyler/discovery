package api_monitor

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ziipin-server/guangmu_go"
	"github.com/ziipin-server/niuhe"
)

type apiMonitor struct {
	watchPath map[string]struct{}
	counter   *prometheus.CounterVec
	latency   *prometheus.HistogramVec
	watchAll  bool
}

func MonitorAPI(pc guangmu_go.PromConfig, svr *niuhe.Server, watchAll bool, watchPaths []string) {
	buckets := []float64{
		0.5, 1, 3, 5, 10, 15, 20, 30, 40, 50, 75, 100, 150, 200, 400, 700, 1000, 2000, 3000, 5000, 10000}
	watchPathSet := make(map[string]struct{})
	for _, watchPath := range watchPaths {
		watchPathSet[watchPath] = struct{}{}
	}
	am := &apiMonitor{
		watchPath: watchPathSet,
		watchAll:  watchAll,
		counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "module_responses",
				Help: "used to calculate qps and failure ratio",
			},
			[]string{"api", "method", "code"},
		),
		latency: prometheus.NewHistogramVec( // P95/P99
			prometheus.HistogramOpts{
				Name:    "response_duration_milliseconds",
				Help:    "HTTP latency distributions",
				Buckets: buckets,
			},
			[]string{"api", "method", "code"},
		),
	}
	pc.MustRegister(am.counter, am.latency)
	svr.Use(prometheusMiddlewareHandler(am))
}

func prometheusMiddlewareHandler(cfg *apiMonitor) gin.HandlerFunc {
	return func(context *gin.Context) {
		start := time.Now()
		context.Next()

		status := context.Writer.Status()
		if cfg == nil || status == 404 {
			return
		}
		if !cfg.watchAll {
			if _, ok := cfg.watchPath[context.Request.URL.Path]; !ok {
				return
			}
		}
		d := time.Since(start)
		cfg.LogQuery(context.Request.URL.Path, context.Request.Method, status, d)
	}
}

func (am *apiMonitor) LogQuery(api, method string, code int, dur time.Duration) {
	codeStr := strconv.Itoa(code)
	am.counter.WithLabelValues(
		api,
		method,
		codeStr,
	).Add(1)

	mSec := dur / time.Millisecond
	nSec := dur % time.Millisecond
	latency := float64(mSec) + float64(nSec)/1e6
	am.latency.WithLabelValues(
		api,
		method,
		codeStr,
	).Observe(latency)
}
