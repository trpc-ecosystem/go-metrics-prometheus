package prometheus

import (
	"strings"

	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/plugin"
)

const (
	pluginType = "metrics"
	pluginName = "prometheus"
)

func init() {
	//register plugin.
	plugin.Register(pluginName, &Plugin{})
	//register filter.
	filter.Register(pluginName, ServerFilter, ClientFilter)
}

// Config config struct.
type Config struct {
	IP           string `yaml:"ip"`           //metrics monitoring address.
	Port         int32  `yaml:"port"`         //metrics listens to the port.
	Path         string `yaml:"path"`         //metrics path.
	Namespace    string `yaml:"namespace"`    //formal or test.
	Subsystem    string `yaml:"subsystem"`    //default trpc.
	RawMode      bool   `yaml:"rawmode"`      //by default, the special character in metrics will be converted.
	EnablePush   bool   `yaml:"enablepush"`   //push is not enabled by default.
	Password     string `yaml:"password"`     //account Password.
	Gateway      string `yaml:"gateway"`      //push gateway address.
	PushInterval uint32 `yaml:"pushinterval"` //push interval,default 1s.
	Job          string `yaml:"job"`          //reported task name.
}

// Default set default values
func (c Config) Default() *Config {
	return &Config{
		IP:           "127.0.0.1",
		Port:         8080,
		Path:         "/metrics",
		Namespace:    "Development",
		Subsystem:    "trpc",
		RawMode:      false,
		EnablePush:   false,
		Gateway:      "",
		PushInterval: 1,
		Job:          "",
	}
}

// Plugin plugin obj
type Plugin struct {
}

// Type plugin type
func (p *Plugin) Type() string {
	return pluginType
}

// Setup init plugin
func (p *Plugin) Setup(name string, decoder plugin.Decoder) error {

	cfg := Config{}.Default()

	err := decoder.Decode(cfg)
	if err != nil {
		log.Errorf("trpc-metrics-prometheus:conf Decode error:%v", err)
		return err
	}
	go func() {
		err := initMetrics(cfg.IP, cfg.Port, cfg.Path)
		if err != nil {
			log.Errorf("trpc-metrics-prometheus:running:%v", err)
		}
	}()
	initSink(cfg)

	return nil
}

func basicAuthForPasswordOption(s string) (username, password string) {
	splits := strings.Split(s, ":")
	if len(splits) < 2 {
		return
	}

	return splits[0], splits[1]
}
