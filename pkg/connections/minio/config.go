package minio

type Config struct {
	Endpoint    string `yaml:"endpoint"`
	AccessKey   string `yaml:"access_key"`
	SecretKey   string `yaml:"secret_key"`
	SecretToken string `yaml:"secret_token"`
	UseSSL      bool   `yaml:"use_ssl"`
}
