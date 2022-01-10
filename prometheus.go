package guangmu_go

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// ** 通用Metrics start **
type PromConfig struct {
	serviceName string
	host        string
	push        *prometheus.CounterVec
	constLabels []*io_prometheus_client.LabelPair
	register    prometheus.Registerer
	gather      prometheus.Gatherer
}

func NewCollecter(serviceName string, reg prometheus.Registerer, gather prometheus.Gatherer) *PromConfig {
	if strings.TrimSpace(serviceName) == "" {
		panic(serviceName + " is empty")
	}
	config := &PromConfig{
		serviceName: serviceName,
		host:        GetHost(),
		push:        prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"state"}),
		register:    reg,
		gather:      gather,
	}
	config.register.MustRegister(config.push)
	labels := []string{"job", serviceName, "instance", config.host}
	config.constLabels = []*io_prometheus_client.LabelPair{
		{Name: &labels[0], Value: &labels[1]},
		{Name: &labels[2], Value: &labels[3]},
	}
	return config
}

func (pc *PromConfig) BackgroundPush(pushAddr string, pushIntervalSec int) {
	if strings.TrimSpace(pushAddr) == "" {
		panic("pushAddr is empty")
	}
	if pushIntervalSec <= 0 {
		pushIntervalSec = 60
	}
	go func() {
		t := time.NewTicker(time.Duration(pushIntervalSec) * time.Second)
		defer t.Stop()
		for {
			<-t.C
			pc.PushMetrics(pushAddr)
		}
	}()
}

func (pc *PromConfig) MustRegister(collectors ...prometheus.Collector) {
	pc.register.MustRegister(collectors...)
}

func (pc *PromConfig) PushMetrics(pushAddr string) {
	mfs, err := pc.gather.Gather()
	if err != nil {
		pc.push.WithLabelValues("gather-error").Inc()
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
			return
		}
	}

	err = DoPush(pushAddr, pc.serviceName, pc.host, buf.String())
	if err != nil {
		pc.push.WithLabelValues("push failed").Inc()
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

// ** 通用Metrics end **

// ** 自定义模块(不在niuhe里注册的那部分api)使用示例 start **
// 延迟
// latencyMS := endTime - startTime // your_api_lantency
// prometheus.GetMonitorWrapper().LogQuery("/api/account/inner_check_token/", "POST", 200, latencyMS)
// ** 自定义模块 end **
// **如有更多的需求，可以扩展
