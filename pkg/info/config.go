package info

type Config struct {
	Hostname string
	AppName  string
}

func (c *Config) BuildVersion() string {
	if CommitTag != "" {
		return CommitTag
	}

	return CommitShortSHA
}
