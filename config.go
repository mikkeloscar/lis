package lis

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Config defines the lis config struct.
type Config struct {
	StateFile string `toml:"statefile"`
	Backlight string `toml:"backlight"`
	IdleTime  uint   `toml:"idle"`
}

// ReadConfig reads the config from filePath.
func ReadConfig(filePath string) (*Config, error) {
	var conf Config
	_, err := toml.DecodeFile(filePath, &conf)
	if err != nil {
		return nil, err
	}

	if conf.Backlight != "intel" {
		return nil, fmt.Errorf("invalid backlight type: %s", conf.Backlight)
	}

	return &conf, nil
}
