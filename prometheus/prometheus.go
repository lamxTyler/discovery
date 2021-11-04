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
	pushAddr    string
	watchPath   map[string]struct{}
	counter     *prometheus.CounterVec
	latency     *prometheus.HistogramVec
	push        *prometheus.CounterVec
	constLabels []*io_prometheus_client.LabelPair
	watchAll    bool
	register    prometheus.Registerer
	gather      prometheus.Gatherer
}

var defaultConfig *promConfig

func GetMonitorWrapper() *promConfig {
	return defaultConfig
}

func InitMonitoring(svr *niuhe.Server, pushIntervalSec int, pushAddr, serviceName string, watchAll bool, watchPaths []string) {
	if strings.TrimSpace(serviceName) == "" || strings.TrimSpace(pushAddr) == "" {
		panic(serviceName + " or pushAddr is empty !!!")
	}
	buckets := []float64{
		0.5, 1, 3, 5, 10, 15, 20, 30, 40, 50, 75, 100, 150, 200, 400, 700, 1000, 2000, 3000, 5000, 10000}
	watchPathSet := make(map[string]struct{})
	for _, watchPath := range watchPaths {
		watchPathSet[watchPath] = struct{}{}
	}

	defaultConfig = &promConfig{
		serviceName: serviceName,
		host:        host.GetHost(),
		pushAddr:    strings.TrimSpace(pushAddr) + "/api/metrics/add/",
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
			[]string{"api", "method", "code"},
		),
		push: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "prom_metric_push_total",
				Help: "Total number of pushes.",
			},
			[]string{"state"},
		),
		watchAll: watchAll,
		register: prometheus.DefaultRegisterer,
		gather:   prometheus.DefaultGatherer,
	}
	labels := []string{"job", serviceName, "instance", defaultConfig.host}
	defaultConfig.constLabels = []*io_prometheus_client.LabelPair{
		{Name: &labels[0], Value: &labels[1]},
		{Name: &labels[2], Value: &labels[3]},
	}
	defaultConfig.register.MustRegister(defaultConfig.counter)
	defaultConfig.register.MustRegister(defaultConfig.latency)
	defaultConfig.register.MustRegister(defaultConfig.push)
	go func() {
		if pushIntervalSec <= 0 {
			pushIntervalSec = 60
		}
		t := time.NewTicker(time.Duration(pushIntervalSec) * time.Second)
		defer t.Stop()
		for {
			<-t.C
			defaultConfig.PushMetrics()
		}
	}()

	svr.Use(prometheusMiddlewareHandler())
}

func NewCollecter(pushAddr, serviceName, host string, registry *prometheus.Registry) *promConfig {
	config := &promConfig{
		serviceName: serviceName,
		host:        host,
		pushAddr:    strings.TrimSpace(pushAddr) + "/api/metrics/add/",
		push:        prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"state"}),
		register:    registry,
		gather:      registry,
	}
	labels := []string{"job", serviceName, "instance", config.host}
	config.constLabels = []*io_prometheus_client.LabelPair{
		{Name: &labels[0], Value: &labels[1]},
		{Name: &labels[2], Value: &labels[3]},
	}
	return config
}

func (pc *promConfig) LogQuery(api, method string, code int, dur time.Duration) {
	codeStr := strconv.Itoa(code)
	pc.counter.WithLabelValues(
		api,
		method,
		codeStr,
	).Add(1)

	mSec := dur / time.Millisecond
	nSec := dur % time.Millisecond
	latency := float64(mSec) + float64(nSec)/1e6
	pc.latency.WithLabelValues(
		api,
		method,
		codeStr,
	).Observe(latency)
}

func (pc *promConfig) PushMetrics() {
	mfs, err := pc.gather.Gather()
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
			m.Label = append(m.Label, pc.constLabels...)
			m.TimestampMs = &timeStamp
		}
		err := enc.Encode(mf)
		if err != nil {
			pc.push.WithLabelValues("encode-error").Inc()
			niuhe.LogError("encode metrics failed: %s", err.Error())
			return
		}
	}

	err = DoPush(pc.pushAddr, pc.serviceName, pc.host, buf.String())
	if err != nil {
		pc.push.WithLabelValues("push failed").Inc()
		niuhe.LogError("push failed: %s", err.Error())
		return
	}
	pc.push.WithLabelValues("success").Inc()
}

func DoPush(pushAddr, serviceName, instance, metricsData string) error {
	values := map[string]string{
		"service_name":  serviceName,
		"instance":      instance,
		"metrics_datas": metricsData,
	}
	data, err := json.Marshal(values)
	if err != nil {
		return err
	}
	resp, err := http.Post(pushAddr, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func BuildLabelPairs(labels []string) []*io_prometheus_client.LabelPair {
	pairs := make([]*io_prometheus_client.LabelPair, 0, len(labels)/2)
	for i := 0; i < len(labels); i += 2 {
		pairs = append(pairs, &io_prometheus_client.LabelPair{
			Name:  &labels[i],
			Value: &labels[i+1],
		})
	}
	return pairs
}

func WriteGuageMetrics(encoder expfmt.Encoder, name string, labels []string, value float64, ts int64) error {
	tp := io_prometheus_client.MetricType_GAUGE
	mf := &io_prometheus_client.MetricFamily{
		Name: &name,
		Help: &name,
		Type: &tp,
		Metric: []*io_prometheus_client.Metric{
			{
				Label:       BuildLabelPairs(labels),
				Gauge:       &io_prometheus_client.Gauge{Value: &value},
				TimestampMs: &ts,
			},
		},
	}
	return encoder.Encode(mf)
}

func prometheusMiddlewareHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		start := time.Now()
		context.Next()

		status := context.Writer.Status()
		if defaultConfig == nil || status == 404 {
			return
		}
		if !defaultConfig.watchAll {
			if _, ok := defaultConfig.watchPath[context.Request.URL.Path]; !ok {
				return
			}
		}
		d := time.Since(start)
		defaultConfig.LogQuery(context.Request.URL.Path, context.Request.Method, status, d)
	}
}

// ** 通用Metrics end **

// ** 自定义模块(不在niuhe里注册的那部分api)使用示例 start **
// 延迟
// latencyMS := endTime - startTime // your_api_lantency
// prometheus.GetMonitorWrapper().LogQuery("/api/account/inner_check_token/", "POST", 200, latencyMS)
// ** 自定义模块 end **
// **如有更多的需求，可以扩展
