package prometheus

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus/push"

	"trpc.group/trpc-go/trpc-go/metrics"
	runtime "trpc.group/trpc-go/trpc-metrics-runtime"
)

func TestSink(t *testing.T) {
	setup(t)
	s := Sink{
		enablePush: true,
		pusher:     push.New("", ""),
	}
	i := float64(0)
	for i <= 3 {
		s.incrCounter("test_counter", 100*i)
		s.addSample("test_sample", 200*i)
		s.setGauge("test_gauge", 300*i)
		_ = s.Report(metrics.NewSingleDimensionMetrics("test_counter_中文", 1, metrics.PolicySUM))

		// Test multi-dimensional, multi-record reporting.
		for _, policy := range []metrics.Policy{metrics.PolicySUM, metrics.PolicySET, metrics.PolicyHistogram} {
			// report labels from 0 to 20
			ms := make([]*metrics.Metrics, 0)
			labels := make([]*metrics.Dimension, 0)
			for j := 0; j < 20; j++ {
				labels = append(labels, &metrics.Dimension{Name: fmt.Sprintf("test_labels_%d", j), Value: fmt.Sprintf("test_value_%d", int(i))})
				ms = append(ms, metrics.NewMetrics(fmt.Sprintf("test_counter_%d", int(i)), float64(100*j), policy))
				_ = s.Report(metrics.NewMultiDimensionMetricsX(fmt.Sprintf("test_name_%d_%d", j, policy), labels, ms))
				_ = metrics.ReportMultiDimensionMetricsX(fmt.Sprintf("test_name_%d_%d", j, policy), labels, ms)
			}
		}
		i++
	}

	// Test multi-dimensional, single-record reporting.
	ms := make([]*metrics.Metrics, 0)
	labels := make([]*metrics.Dimension, 0)
	labels = append(labels, &metrics.Dimension{Name: fmt.Sprintf("test_labels_%d", 99), Value: fmt.Sprintf("test_value_%d", int(99))})
	ms = append(ms, metrics.NewMetrics(fmt.Sprintf("test_counter_%d", int(99)), float64(100*99), metrics.PolicySUM))
	_ = metrics.ReportMultiDimensionMetricsX("", labels, ms)

	t.Log(getMetrics(t))
}

func TestMetrics(t *testing.T) {
	setup(t)
	s := Sink{
		enablePush: true,
		pusher:     push.New("", ""),
	}
	bu := metrics.NewValueBounds(100, 500, 800)
	metrics.RegisterMetricsSink(&s)
	i := float64(0)
	for i <= 10 {
		metrics.AddSample("test_sample", bu, 100*i)
		i++
	}
	t.Log(getMetrics(t))
}

func TestRuntime(t *testing.T) {
	setup(t)
	cfg := &Config{Namespace: "test", Subsystem: "testing", RawMode: false, EnablePush: true, PushInterval: 1}
	initSink(cfg)
	GetDefaultPusher()
	GetDefaultPrometheusSink()
	//metrics.RegisterMetricsSink(s)
	runtime.RuntimeMetrics()
	t.Log(getMetrics(t))
}
