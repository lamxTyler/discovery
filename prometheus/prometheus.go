package prometheus

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lamxTyler/discovery/utils/host"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/ziipin-server/niuhe"
)

// ** 通用Metrics start **
type promConfig struct {
	serviceName string
	host        string
	watchPath   map[string]struct{}
	counter     *prometheus.CounterVec
	latency     *prometheus.HistogramVec
	push        *prometheus.CounterVec
	constLabels []*io_prometheus_client.LabelPair
}

var config *promConfig

func GetMonitorWrapper() *promConfig {
	return config
}

func InitMonitoring(svr *niuhe.Server, pushIntervalSec int, pushAddr, serviceName string, watchPaths []string) {
	if strings.TrimSpace(serviceName) == "" || strings.TrimSpace(pushAddr) == "" {
		panic(serviceName + " or pushAddr is empty !!!")
	}
	buckets := []float64{
		0.5, 1, 3, 5, 10, 15, 20, 30, 40, 50, 75, 100, 150, 200, 400, 700, 1000, 2000, 3000, 5000, 10000}
	watchPathSet := make(map[string]struct{})
	for _, watchPath := range watchPaths {
		watchPathSet[watchPath] = struct{}{}
	}

	config = &promConfig{
		serviceName: serviceName,
		host:        host.GetHost(),
		watchPath:   watchPathSet,
		counter: prometheus.NewCounterVec( // QPS and failure ratio
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
			[]string{"api", "method"},
		),
		push: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "prom_metric_push_total",
				Help: "Total number of pushes.",
			},
			[]string{"state"},
		),
	}
	labels := []string{"job", serviceName, "instance", config.host}
	config.constLabels = []*io_prometheus_client.LabelPair{
		{Name: &labels[0], Value: &labels[1]},
		{Name: &labels[2], Value: &labels[3]},
	}
	prometheus.MustRegister(config.counter)
	prometheus.MustRegister(config.latency)
	prometheus.MustRegister(config.push)
	go func() {
		if pushIntervalSec <= 0 {
			pushIntervalSec = 60
		}
		t := time.NewTicker(time.Duration(pushIntervalSec) * time.Second)
		addr := strings.TrimSpace(pushAddr) + "/api/metrics/add/"
		defer t.Stop()
		for {
			<-t.C
			config.pushMetrics(addr)
		}
	}()

	svr.Use(prometheusMiddlewareHandler())
}

func (pc *promConfig) LogQuery(api, method string, code int, dur time.Duration) {
	pc.counter.WithLabelValues(
		pc.serviceName,
		api,
		method,
		strconv.Itoa(code),
	).Add(1)

	mSec := dur / time.Millisecond
	nSec := dur % time.Millisecond
	latency := float64(mSec) + float64(nSec)/1e6
	pc.latency.WithLabelValues(
		pc.serviceName,
		api,
		method,
	).Observe(latency)
}

func (pc *promConfig) pushMetrics(pushAddr string) {
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		pc.push.WithLabelValues("gather-error").Inc()
		niuhe.LogError("gather metrics failed: %s", err.Error())
		return
	}
	buf := &bytes.Buffer{}
	enc := expfmt.NewEncoder(buf, expfmt.FmtText)
	now := time.Now()
	timeStamp := now.Unix()*1000 + int64(now.Nanosecond()/1000000)
	for _, mf := range mfs {
		for _, m := range mf.Metric {
			m.Label = append(m.Label, config.constLabels...)
			m.TimestampMs = &timeStamp
		}
		err := enc.Encode(mf)
		if err != nil {
			pc.push.WithLabelValues("encode-error").Inc()
			niuhe.LogError("encode metrics failed: %s", err.Error())
			return
		}
	}

	values := map[string]string{
		"service_name":  pc.serviceName,
		"instance":      pc.host,
		"metrics_datas": buf.String(),
	}
	data, err := json.Marshal(values)
	if err != nil {
		pc.push.WithLabelValues("marshal-error").Inc()
		niuhe.LogError("marshal json failed: %s", err.Error())
		return
	}
	resp, err := http.Post(pushAddr, "application/json", bytes.NewBuffer(data))
	if err != nil {
		pc.push.WithLabelValues("push-error").Inc()
		niuhe.LogError("push metrics failed: %s", err.Error())
		return
	}
	resp.Body.Close()
	pc.push.WithLabelValues("success").Inc()
}

func prometheusMiddlewareHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		start := time.Now()
		context.Next()

		if config == nil {
			return
		}
		if _, ok := config.watchPath[context.Request.URL.Path]; !ok {
			return
		}
		d := time.Since(start)
		config.LogQuery(context.Request.URL.Path, context.Request.Method, context.Writer.Status(), d)
	}
}

// ** 通用Metrics end **

// ** 自定义模块(不在niuhe里注册的那部分api)使用示例 start **
// 延迟
// latencyMS := endTime - startTime // your_api_lantency
// prometheus.GetMonitorWrapper().LogQuery("/api/account/inner_check_token/", "POST", 200, latencyMS)
// ** 自定义模块 end **
// **如有更多的需求，可以扩展
