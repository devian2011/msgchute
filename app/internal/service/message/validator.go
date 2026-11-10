package message

import (
	"fmt"

	"github.com/devian2011/msgchute/internal/dto"
	"github.com/devian2011/msgchute/pkg/helper"
)

type ProviderMapper interface {
	GetProviders() []string
}

type Validator struct {
	providerMapper ProviderMapper
}

func NewValidator(providerMapper ProviderMapper) *Validator {
	return &Validator{
		providerMapper: providerMapper,
	}
}

func (v *Validator) Validate(msg *dto.Message) (bool, map[string]error) {
	if msg == nil {
		return true, map[string]error{"message": fmt.Errorf("empty message")}
	}
	result := make(map[string]error)
	providersMap := v.providerMapper.GetProviders()
	if !helper.InArray(msg.Transport, providersMap) {
		result["transport"] = fmt.Errorf("unknown transport: %s", msg.Transport)
	}

	return len(result) > 0, result
}
