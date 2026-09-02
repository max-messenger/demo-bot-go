package bigbro

type Config struct {
	Enabled     bool           `yaml:"enabled"`
	Token       string         `yaml:"token"`
	URL         string         `yaml:"url"`
	ExtraFields map[string]any `yaml:"extra_fields"`
}
