package prometheus

import (
	"context"
	"fmt"
	"strings"
	"time"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/metrics"
)

// Set the time consumption interval.
// 1 10 20 40 80 160 1000 2000 3000
var clientBounds = metrics.NewValueBounds(1.0, 10.0, 20.0, 40.0, 80.0, 160.0, 1000.0, 2000.0, 3000.0)
var serverBounds = clientBounds

// SetClientBounds set client filter buckets.
func SetClientBounds(b metrics.BucketBounds) {
	clientBounds = b
}

// SetServerBounds set server filter buckets.
func SetServerBounds(b metrics.BucketBounds) {
	serverBounds = b
}

func getLabels(msg codec.Msg, err error) []*metrics.Dimension {
	code := fmt.Sprintf("%d", errs.RetOK)
	if err != nil {
		e, ok := err.(*errs.Error)
		if ok && e != nil {
			if e.Desc != "" {
				code = fmt.Sprintf("%s_%d", e.Desc, e.Code)
			} else {
				code = fmt.Sprintf("%d", e.Code)
			}
		} else {
			code = fmt.Sprintf("%d", errs.RetUnknown)
		}
	}
	var remoteAddr, localAddr string
	if msg.RemoteAddr() != nil {
		remoteAddr = getAddr(msg.RemoteAddr().String())
	}
	if msg.LocalAddr() != nil {
		localAddr = getAddr(msg.LocalAddr().String())
	}
	return []*metrics.Dimension{
		{Name: "CallerService", Value: msg.CallerService()},
		{Name: "CallerMethod", Value: msg.CallerMethod()},
		{Name: "CalleeService", Value: msg.CalleeService()},
		{Name: "CalleeMethod", Value: msg.CalleeMethod()},
		{Name: "CalleeContainerName", Value: msg.CalleeContainerName()},
		{Name: "CalleeSetName", Value: msg.CalleeSetName()},
		{Name: "RemoteAddr", Value: remoteAddr},
		{Name: "LocalAddr", Value: localAddr},
		{Name: "Code", Value: code},
	}
}

// ClientFilter client filter for prome.
func ClientFilter(ctx context.Context, req, rsp interface{}, handler filter.ClientHandleFunc) error {
	begin := time.Now()
	hErr := handler(ctx, req, rsp)
	msg := trpc.Message(ctx)
	labels := getLabels(msg, hErr)
	ms := make([]*metrics.Metrics, 0)
	t := float64(time.Since(begin)) / float64(time.Millisecond)
	ms = append(ms,
		metrics.NewMetrics("time", t, metrics.PolicyHistogram),
		metrics.NewMetrics("requests", 1.0, metrics.PolicySUM))
	metrics.Histogram("ClientFilter_time", clientBounds)
	r := metrics.NewMultiDimensionMetricsX("ClientFilter", labels, ms)
	_ = GetDefaultPrometheusSink().Report(r)
	return hErr
}

// ServerFilter server filter for prome.
func ServerFilter(ctx context.Context, req interface{}, handler filter.ServerHandleFunc) (rsp interface{}, err error) {
	begin := time.Now()
	rsp, err = handler(ctx, req)
	msg := trpc.Message(ctx)
	labels := getLabels(msg, err)
	ms := make([]*metrics.Metrics, 0)
	t := float64(time.Since(begin)) / float64(time.Millisecond)
	ms = append(ms,
		metrics.NewMetrics("time", t, metrics.PolicyHistogram),
		metrics.NewMetrics("requests", 1.0, metrics.PolicySUM))
	metrics.Histogram("ServerFilter_time", serverBounds)
	r := metrics.NewMultiDimensionMetricsX("ServerFilter", labels, ms)
	_ = GetDefaultPrometheusSink().Report(r)
	return rsp, err
}

// getAddr obtains IP, excluding ports.
func getAddr(add string) string {
	var addr string
	s := strings.Split(add, ":")
	if len(s) > 0 {
		addr = s[0]
	}
	return addr
}
