package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	config := NewConfig()

	// 读取配置文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 从环境变量覆盖配置
	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		config.Telegram.BotToken = token
	}

	return config, nil
}

// Validate 验证配置是否有效
func (c *Config) Validate() error {
	if c.Telegram.BotToken == "" {
		return fmt.Errorf("未设置Telegram Bot Token")
	}
	if c.Telegram.TargetGroupID == 0 {
		return fmt.Errorf("未设置目标群组ID")
	}

	// 如果HTTP和WS都启用且端口相同，检查TLS配置是否一致
	if c.HTTP.Enabled && c.WS.Enabled && c.HTTP.Port == c.WS.Port {
		if c.HTTP.TLS.Cert != c.WS.TLS.Cert || c.HTTP.TLS.Key != c.WS.TLS.Key {
			return fmt.Errorf("HTTP和WebSocket使用相同端口时必须使用相同的TLS配置")
		}
	}

	return nil
}
