// Package prometheus is a tRPC metric filter that provides prometheus exporter, Sink and Labels feature.
package prometheus

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	"trpc.group/trpc-go/trpc-go/log"
)

type metricsNameCache struct {
	locker sync.RWMutex
	cache  map[string]string
}

func newMetricsNameCache() *metricsNameCache {
	return &metricsNameCache{
		locker: sync.RWMutex{},
		cache:  make(map[string]string),
	}
}

var (
	mNameCache = newMetricsNameCache()
)

// initMetrics initialize metrics and metrics handler
func initMetrics(ip string, port int32, path string) error {
	metricsHTTPHandler := http.NewServeMux()
	metricsHTTPHandler.Handle(path, promhttp.Handler())
	addr := fmt.Sprintf("%s:%d", ip, port)
	server := &http.Server{
		Addr:    addr,
		Handler: metricsHTTPHandler,
	}
	log.Infof("prometheus exporter running at %s, metrics path %s", addr, path)
	return server.ListenAndServe()
}

// convertSpecialChars convert utf8 chars to _
func convertSpecialChars(in string) (out string) {
	if len(in) == 0 {
		return
	}
	var outBuilder strings.Builder
	for i, b := range []rune(in) {
		if !isNormalChar(i, b) {
			// use '_' instead of illegal character
			if b < 127 {
				outBuilder.WriteByte('_')
			} else {
				// convert utf8 to int string
				outBuilder.WriteString(strconv.Itoa(int(b)))
				outBuilder.WriteByte('_')
			}
			continue
		}
		outBuilder.WriteByte(byte(b))
	}
	return outBuilder.String()
}

// convertSpecialCharsWithCache convert utf8 chars to _ with cache
func convertSpecialCharsWithCache(in string) (out string) {
	if len(in) == 0 {
		return
	}
	// use cache
	mNameCache.locker.RLock()
	if v, ok := mNameCache.cache[in]; ok {
		mNameCache.locker.RUnlock()
		return v
	}
	mNameCache.locker.RUnlock()
	// add cache
	out = convertSpecialChars(in)
	mNameCache.locker.Lock()
	mNameCache.cache[in] = out
	mNameCache.locker.Unlock()
	return
}

func isNum(b rune) bool {
	return b >= '0' && b <= '9'
}

func isChar(b rune) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isNormalChar(i int, b rune) bool {
	return isChar(b) || b == '_' || b == ':' || (isNum(b) && i > 0)
}

// checkMetricsValid metrics only support ascii letters and digists
func checkMetricsValid(name string) bool {
	return model.IsValidMetricName(model.LabelValue(name))
}
