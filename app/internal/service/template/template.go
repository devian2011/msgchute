package template

import (
	"github.com/devian2011/msgchute/internal/dto"
)

type Store interface {
	GetList() ([]*dto.Template, error)
	GetByCode(code string) (*dto.Template, error)
	Save(msg *dto.Template) error
	Delete(code string) error
}
