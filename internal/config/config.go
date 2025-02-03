package config

type Config struct {
	HTTP     HTTPConfig      `yaml:"http"`
	WS       WebSocketConfig `yaml:"ws"`
	Telegram TelegramConfig  `yaml:"telegram"`
}

type HTTPConfig struct {
	Enabled   bool      `yaml:"enabled"`
	Port      int       `yaml:"port"`
	SecretKey string    `yaml:"secret_key"`
	TLS       TLSConfig `yaml:"tls"`
}

type WebSocketConfig struct {
	Enabled   bool      `yaml:"enabled"`
	Port      int       `yaml:"port"`
	SecretKey string    `yaml:"secret_key"`
	TLS       TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type TelegramConfig struct {
	BotToken      string `yaml:"botToken"`
	TargetGroupID int64  `yaml:"targetGroupId"`
}

// NewConfig 返回默认配置
func NewConfig() *Config {
	return &Config{
		HTTP: HTTPConfig{
			Enabled: true,
			Port:    8080,
		},
		WS: WebSocketConfig{
			Enabled: true,
			Port:    8080,
		},
		Telegram: TelegramConfig{},
	}
}
