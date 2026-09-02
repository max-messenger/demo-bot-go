package demo_bot

type BoxOption func(*Box)

func WithAppName(name string) BoxOption {
	return func(o *Box) {
		if len(name) > 0 {
			o.appName = name
		}
	}
}

func WithConfigFile(configFile string) BoxOption {
	return func(o *Box) {
		if len(o.cfgFile) > 0 {
			o.cfgFile = configFile
		}
	}
}
