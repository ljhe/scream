package logrus

type logConf struct {
	Log LogConfig `yaml:"log"`
}

type LogConfig struct {
	LogName    string `yaml:"logName"`
	LogLevel   uint32 `yaml:"logLevel"`
	MaxSize    int    `yaml:"maxSize"`
	MaxBackups int    `yaml:"maxBackups"`
	MaxAge     int    `yaml:"maxAge"`
	SavePath   string `yaml:"savePath"`
}
