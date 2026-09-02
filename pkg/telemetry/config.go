package telemetry

const (
	collectorEndpointTypeGRPC = "grpc"
	collectorEndpointTypeHTTP = "http"
)

type CollectorConfig struct {
	Enabled       bool              `yaml:"enabled"`
	SamplingRatio float64           `yaml:"sampling_ratio"`
	Endpoint      string            `yaml:"endpoint"`
	EndpointType  string            `yaml:"endpoint_type"`
	URLPath       string            `yaml:"url_path"`
	Headers       map[string]string `yaml:"headers"`
}

type Config struct {
	ServiceName     string          `yaml:"-"`         // set manually
	Hostname        string          `yaml:"-"`         // set manually
	Version         string          `yaml:"-"`         // set manually
	CollectorConfig CollectorConfig `yaml:"collector"` //nolint:tagliatelle
}
