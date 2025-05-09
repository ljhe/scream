package logrus

type logConf struct {
	Log LogConfig `yaml:"log"`
}

type LogConfig struct {
	LogName    string `yaml:"log_name"`
	LogLevel   uint32 `yaml:"log_level"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	SavePath   string `yaml:"save_path"`
}
