package prometheus

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ziipin-server/niuhe"
)

// ** 通用Metrics start **
// 目前通用部分主要覆盖niuhe自动注册的api
const (
	Module = "self"
)

type Options struct {
	ServiceName      string
	Idc              string
	WatchPath        map[string]struct{}
	HistogramBuckets []float64
}

type LatencyRecord struct {
	Time   float64
	Api    string
	Module string
	Method string
	Code   int
}

type QpsRecord struct {
	Times  float64
	Api    string
	Module string
	Method string
	Code   int
}

type Prometheus struct {
	ServiceName string
	Idc         string
	WatchPath   map[string]struct{}
	Counter     *prometheus.CounterVec
	Histogram   *prometheus.HistogramVec
}

var Metrics *Prometheus

func GetMonitorWrapper() *Prometheus {
	return Metrics
}

func InitCommonMonitoring(options Options) {
	if strings.TrimSpace(options.ServiceName) == "" || strings.TrimSpace(options.Idc) == "" || len(options.HistogramBuckets) == 0 {
		panic(options.ServiceName + " or " + options.Idc + " or HistogramBuckets is empty !!!")
	}

	Metrics = &Prometheus{
		ServiceName: options.ServiceName,
		Idc:         options.Idc,
		WatchPath:   options.WatchPath,
		Counter: prometheus.NewCounterVec( // QPS and failure ratio
			prometheus.CounterOpts{
				Name: "module_responses",
				Help: "used to calculate qps and failure ratio",
			},
			[]string{"app", "module", "api", "method", "code", "idc"},
		),
		Histogram: prometheus.NewHistogramVec( // P95/P99
			prometheus.HistogramOpts{
				Name:    "response_duration_milliseconds",
				Help:    "HTTP latency distributions",
				Buckets: options.HistogramBuckets,
			},
			[]string{"app", "module", "api", "method", "idc"},
		),
	}

	prometheus.MustRegister(Metrics.Counter)
	prometheus.MustRegister(Metrics.Histogram)
}

// QPS
func (metrics *Prometheus) QpsCounterLog(record QpsRecord) {
	if strings.TrimSpace(record.Module) == "" {
		record.Module = Module
	}

	metrics.Counter.WithLabelValues(
		metrics.ServiceName,
		record.Module,
		record.Api,
		record.Method,
		strconv.Itoa(record.Code),
		metrics.Idc,
	).Add(record.Times)
}

// P95/P99
func (metrics *Prometheus) LatencyLog(record LatencyRecord) {
	if strings.TrimSpace(record.Module) == "" {
		record.Module = Module
	}

	metrics.Histogram.WithLabelValues(
		metrics.ServiceName,
		record.Module,
		record.Api,
		record.Method,
		metrics.Idc,
	).Observe(record.Time)
}

func PrometheusMiddlewareHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		start := time.Now()
		context.Next()

		go func() {
			if GetMonitorWrapper() != nil {
				if _, ok := Metrics.WatchPath[context.Request.URL.Path]; ok {
					// QPS
					Metrics.QpsCounterLog(QpsRecord{
						Times:  1,
						Api:    context.Request.URL.Path,
						Module: Module,
						Method: context.Request.Method,
						Code:   context.Writer.Status(),
					})

					// P95/P99
					d := time.Since(start)
					mSec := d / time.Millisecond
					nSec := d % time.Millisecond
					latency := float64(mSec) + float64(nSec)/1e6
					Metrics.LatencyLog(LatencyRecord{
						Time:   latency,
						Api:    context.Request.URL.Path,
						Module: Module,
						Method: context.Request.Method,
						Code:   context.Writer.Status(),
					})

					//niuhe.LogInfo("access logger => method %v, path %v, status %v, latency %v", context.Request.Method, context.Request.URL.Path, context.Writer.Status(), latency)
				}
			} else {
				niuhe.LogWarn("need to init prometheus!!!")
			}
		}()
	}
}

// ** 通用Metrics end **

// ** 自定义模块(不在niuhe里注册的那部分api)使用示例 start **
// 延迟
// latencyMS := endTime - startTime // your_api_lantency
// prometheus.GetMonitorWrapper().LatencyLog(prometheus.LatencyRecord{
//  Api:  "/api/account/inner_check_token/",
// 	Module: "uc",
// 	Method: "POST",
//  me: latencyMS
// })

// QPS
// prometheus.GetMonitorWrapper().QpsCountLog(prometheus.QPSRecord{
//    Module: "uc",
//    Times:  float64(1),
//    Api:    "/api/account/inner_check_token/",
//    Method: "POST",
//    Code:   200, // your api's code
// })

// ** 自定义模块 end **

// **如有更多的需求，可以扩展
