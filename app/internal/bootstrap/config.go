package bootstrap

import (
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"github.com/joho/godotenv"

	"github.com/devian2011/msgchute/internal/io/storage"
	"github.com/devian2011/msgchute/internal/io/web"
	"github.com/devian2011/msgchute/internal/service/auth"
	"github.com/devian2011/msgchute/internal/service/sender"
)

func loadConfig(cfgFilePath string) (*Config, error) {
	_ = godotenv.Load()

	config.WithOptions(
		config.ParseEnv,
		config.ParseTime,
		func(opt *config.Options) {
			opt.DecoderConfig.TagName = "yaml"
		},
	)
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
	Http      *web.Config     `yaml:"http"`
	Db        *storage.Config `yaml:"db"`
	Auth      *auth.Config    `yaml:"auth"`
	Providers *sender.Config  `yaml:"providers"`
}
