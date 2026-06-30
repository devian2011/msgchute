package bootstrap

import (
	"context"

	"github.com/devian2011/msgchute/internal/registry"
)

func Boostrap(ctx context.Context, cfgFilePath string) (*registry.AppRegistry, error) {
	_, loadCfgErr := loadConfig(cfgFilePath)
	if loadCfgErr != nil {
		return nil, loadCfgErr
	}

	return nil, nil
}
