package prometheus

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	testCfg = &Config{
		IP:   "127.0.0.1",
		Port: 9090,
		Path: "/metrics",
	}
	once sync.Once
	wg   sync.WaitGroup
)

func setup(t *testing.T) {
	cfg := &yaml.Node{}
	p := &Plugin{}
	_ = p.Setup(pluginName, cfg)
	wg.Add(1)
	go func(w *sync.WaitGroup) {
		w.Done()
		once.Do(
			func() {
				// Waiting for a concurrent process to start
				err := initMetrics(testCfg.IP, testCfg.Port, testCfg.Path)
				if err != nil {
					t.Error(err)
					return
				}
			})
	}(&wg)
}

func getMetrics(t *testing.T) string {
	time.Sleep(time.Second)
	wg.Wait()
	resp, err := http.Get(fmt.Sprintf("http://%s:%d%s", testCfg.IP, testCfg.Port, testCfg.Path))
	if err != nil {
		t.Fatal(err)
		return ""
	}
	if resp.StatusCode != 200 {
		t.Fatal(resp.StatusCode)
		return ""
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return ""
	}
	return string(body)
}

func TestInitMetrics(t *testing.T) {
	setup(t)
	t.Log(getMetrics(t))
}

func TestConvertSpecialChars(t *testing.T) {
	in := "trpc.Chinese Indicators"
	out := convertSpecialChars(in)
	t.Log(out)
	if !checkMetricsValid(out) {
		t.Fatal()
	}
}

func TestConvertSpecialCharsWithCache(t *testing.T) {
	in := "trpc.Chinese Indicators"
	out := convertSpecialCharsWithCache(in)
	out1 := convertSpecialCharsWithCache(in)
	t.Log(out)
	t.Log(out1)
	if !checkMetricsValid(out) || out != out1 {
		t.Fatal()
	}
}

func BenchmarkConvertSpecialChars(b *testing.B) {
	in := "trpc.Chinese Indicators"
	var out string
	for n := 0; n < b.N; n++ {
		out = convertSpecialChars(in)
	}
	if !checkMetricsValid(out) {
		b.Fatal()
	}
}

func BenchmarkConvertSpecialCharsWithCache(b *testing.B) {
	in := "trpc.Chinese Indicators"
	var out string
	for n := 0; n < b.N; n++ {
		out = convertSpecialCharsWithCache(in)
	}
	if !checkMetricsValid(out) {
		b.Fatal()
	}
}
