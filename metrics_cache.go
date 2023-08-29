package prometheus

import (
	"sync"
)

// metricsCache Indicator Cache
type metricsCache struct {
	locker sync.RWMutex
	cache  map[string]interface{}
}

// NewMetricsCache Create Cache
func NewMetricsCache() *metricsCache {
	return &metricsCache{
		locker: sync.RWMutex{},
		cache:  make(map[string]interface{}),
	}
}

// Methods for creating metrics
type createMetricFunc func() interface{}

// Loader loads the indicator from the cache, and if the cache does not exist, execute the create method to create it
func (mc *metricsCache) Loader(key string, f createMetricFunc) interface{} {
	mc.locker.RLock()
	if v, ok := mc.cache[key]; ok {
		mc.locker.RUnlock()
		return v
	}
	mc.locker.RUnlock()

	mc.locker.Lock()
	if v, ok := mc.cache[key]; ok {
		mc.locker.Unlock()
		return v
	}

	// 执行创建函数
	v := f()
	mc.cache[key] = v
	mc.locker.Unlock()

	return v
}
