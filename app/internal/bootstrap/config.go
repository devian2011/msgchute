package bootstrap

import (
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"github.com/joho/godotenv"
)

func loadConfig(cfgFilePath string) (*Config, error) {
	_ = godotenv.Load()

	config.WithOptions(config.ParseEnv)
	config.WithOptions(func(opt *config.Options) {
		opt.DecoderConfig.TagName = "config"
	})
	config.AddDriver(yaml.Driver)
	loadCfgErr := config.LoadFiles(cfgFilePath)
	if loadCfgErr != nil {
		return nil, loadCfgErr
	}

	cfg := &Config{}

	decodeErr := config.Decode(&cfg)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return cfg, nil
}

type Config struct {
}
