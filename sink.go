package prometheus

import (
	"errors"
	"time"

	"github.com/prometheus/client_golang/prometheus/push"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/metrics"
)

const (
	sinkName = "prometheus"
)

var (
	// prometheus only support one metrics registry once.
	// use global cache.
	cache     = NewMetricsCache()
	errLength = errors.New("inconsistent labels and values length")

	// defaultPrometheusPusher default prometheus pusher.
	defaultPrometheusPusher *push.Pusher
	// defaultPrometheusSink default sink for register.
	defaultPrometheusSink *Sink
)

// GetDefaultPusher get default pusher.
func GetDefaultPusher() *push.Pusher {
	return defaultPrometheusPusher
}

// GetDefaultPrometheusSink get prome sink.
func GetDefaultPrometheusSink() *Sink {
	return defaultPrometheusSink
}

func initSink(cfg *Config) {
	defaultPrometheusPusher = push.New(cfg.Gateway, cfg.Job)
	// set basic auth if set.
	if len(cfg.Password) > 0 {
		defaultPrometheusPusher.BasicAuth(basicAuthForPasswordOption(cfg.Password))
	}
	defaultPrometheusSink = &Sink{
		ns:         cfg.Namespace,
		subsystem:  cfg.Subsystem,
		rawMode:    cfg.RawMode,
		enablePush: cfg.EnablePush,
		pusher:     defaultPrometheusPusher,
	}
	metrics.RegisterMetricsSink(defaultPrometheusSink)
	//start up pusher if needed.
	if cfg.EnablePush {
		defaultPrometheusPusher.Gatherer(prometheus.DefaultGatherer)
		go pusherRun(cfg, defaultPrometheusPusher)
	}
}

// start up prometheus pusher.
func pusherRun(cfg *Config, pusher *push.Pusher) {
	ticker := time.NewTicker(time.Duration(cfg.PushInterval) * time.Second)
	for {
		<-ticker.C
		err := pusher.Push()
		if err != nil {
			log.Errorf("push result=%v", err)
		}
	}
}

// Sink struct
type Sink struct {
	//ns namespace for metrics.
	ns string
	//subsystem ns.
	subsystem string
	//rawMode convert special char metrics.
	rawMode bool
	//enable push.
	enablePush bool
	//Pusher manages a push to the pushgateway.
	pusher *push.Pusher
}

// Name return sink name.
func (s *Sink) Name() string {
	return sinkName
}

// GetMetricsName returns metrics name.
func (s *Sink) GetMetricsName(m *metrics.Metrics) string {
	if !s.rawMode {
		return convertSpecialCharsWithCache(m.Name())
	}

	return m.Name()
}

// Report report.
func (s *Sink) Report(rec metrics.Record, opts ...metrics.Option) error {
	if len(rec.GetDimensions()) <= 0 {
		return s.ReportSingleLabel(rec, opts...)
	}
	labels := make([]string, 0)
	values := make([]string, 0)
	prefix := rec.GetName()

	if len(labels) != len(values) {
		return errLength
	}

	for _, dimension := range rec.GetDimensions() {
		labels = append(labels, dimension.Name)
		values = append(values, dimension.Value)
	}
	for _, m := range rec.GetMetrics() {
		name := s.GetMetricsName(m)
		if prefix != "" {
			name = prefix + "_" + name
		}
		if !checkMetricsValid(name) {
			log.Errorf("metrics %s(%s) is invalid", name, m.Name())
			continue
		}
		s.reportVec(name, m, labels, values)
	}
	return nil
}

func (s *Sink) reportVec(name string, m *metrics.Metrics, labels, values []string) {
	switch m.Policy() {
	case metrics.PolicySUM:
		s.incrCounterVec(name, m.Value(), labels, values)
	case metrics.PolicySET:
		s.setGaugeVec(name, m.Value(), labels, values)
	case metrics.PolicyHistogram:
		s.addSampleVec(name, m.Value(), labels, values)
	default:
		log.Warnf("trpc-metrics-prometheus Policy not support %d", m.Policy())
	}
}

// ReportSingleLabel single indicator report.
func (s *Sink) ReportSingleLabel(rec metrics.Record, opts ...metrics.Option) error {
	for _, m := range rec.GetMetrics() {
		name := s.GetMetricsName(m)
		if !checkMetricsValid(name) {
			log.Errorf("metrics %s(%s) is invalid", name, m.Name())
			continue
		}
		s.report(name, m)
	}
	return nil
}

func (s *Sink) report(name string, m *metrics.Metrics) {
	switch m.Policy() {
	case metrics.PolicySUM:
		s.incrCounter(name, m.Value())
	case metrics.PolicySET:
		s.setGauge(name, m.Value())
	case metrics.PolicyHistogram:
		s.addSample(name, m.Value())
	default:
		log.Warnf("trpc-metrics-prometheus Policy not support %d", m.Policy())
	}
}

func (s *Sink) incrCounterVec(key string, value float64, labels []string, values []string) {
	cacheKey := "countervec_" + key
	v := cache.Loader(cacheKey, func() interface{} {
		// Create metrics.
		return promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: s.ns,
			Subsystem: s.subsystem,
			Name:      key,
		}, labels)
	})

	counterVec := v.(*prometheus.CounterVec)
	counterVec.WithLabelValues(values...).Add(value)
}

func (s *Sink) incrCounter(key string, value float64) {
	cacheKey := "counter_" + key
	v := cache.Loader(cacheKey, func() interface{} {
		return promauto.NewCounter(prometheus.CounterOpts{
			Namespace: s.ns,
			Subsystem: s.subsystem,
			Name:      key,
		})
	})

	counter := v.(prometheus.Counter)
	counter.Add(value)
}

func (s *Sink) setGauge(key string, value float64) {
	cacheKey := "gauge_" + key
	v := cache.Loader(cacheKey, func() interface{} {
		return promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: s.ns,
			Subsystem: s.subsystem,
			Name:      key,
		})
	})

	gauge := v.(prometheus.Gauge)
	gauge.Set(value)
}

func (s *Sink) setGaugeVec(key string, value float64, labels []string, values []string) {
	cacheKey := "gaugevec_" + key
	v := cache.Loader(cacheKey, func() interface{} {
		return promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: s.ns,
			Subsystem: s.subsystem,
			Name:      key,
		}, labels)
	})

	gaugeVec := v.(*prometheus.GaugeVec)
	gaugeVec.WithLabelValues(values...).Set(value)
}

func (s *Sink) addSample(key string, value float64) {
	cacheKey := "histogram_" + key
	v := cache.Loader(cacheKey, func() interface{} {
		var buckets []float64
		if h, ok := metrics.GetHistogram(key); ok {
			for _, b := range h.GetBuckets() {
				buckets = append(buckets, b.ValueUpperBound)
			}
		} else {
			buckets = prometheus.DefBuckets
		}
		return promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: s.ns,
			Subsystem: s.subsystem,
			Name:      key,
			Buckets:   buckets,
		})
	})

	histogram := v.(prometheus.Histogram)
	histogram.Observe(value)
}

func (s *Sink) addSampleVec(key string, value float64, labels []string, values []string) {
	cacheKey := "histogramvec_" + key
	v := cache.Loader(cacheKey, func() interface{} {
		var buckets []float64
		if h, ok := metrics.GetHistogram(key); ok {
			for _, b := range h.GetBuckets() {
				buckets = append(buckets, b.ValueUpperBound)
			}
		} else {
			buckets = prometheus.DefBuckets
		}
		return promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: s.ns,
			Subsystem: s.subsystem,
			Name:      key,
			Buckets:   buckets,
		}, labels)
	})

	histogramVec := v.(*prometheus.HistogramVec)
	histogramVec.WithLabelValues(values...).Observe(value)
}
