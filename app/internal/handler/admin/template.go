package admin

import (
	"github.com/devian2011/msgchute/internal/dto"
)

// Create handler

type templateCreator interface {
	Create(*dto.Template) (*dto.Template, error)
}

type TemplateCreateHandler struct {
	creator templateCreator
}

func NewTemplateCreateHandler(creator templateCreator) *TemplateCreateHandler {
	return &TemplateCreateHandler{creator: creator}
}

func (h *TemplateCreateHandler) Handle(t *dto.Template) (*dto.Template, error) {
	return h.creator.Create(t)
}

// Update handler

type templateUpdater interface {
	Update(t *dto.Template) (*dto.Template, error)
}

type TemplateUpdateHandler struct {
	updater templateUpdater
}

func NewTemplateUpdateHandler(updater templateUpdater) *TemplateUpdateHandler {
	return &TemplateUpdateHandler{updater: updater}
}

func (h *TemplateUpdateHandler) Handle(t *dto.Template) (*dto.Template, error) {
	return h.updater.Update(t)
}

// Find handler

type templateFinder interface {
	Find(filter *dto.MessageTemplateFilter) (map[string]*dto.Template, uint64, error)
}

type TemplateFinderHandler struct {
	finder templateFinder
}

func NewTemplateFinderHandler(finder templateFinder) *TemplateFinderHandler {
	return &TemplateFinderHandler{finder: finder}
}

func (h *TemplateFinderHandler) Handle(filter *dto.MessageTemplateFilter) (map[string]*dto.Template, uint64, error) {
	return h.finder.Find(filter)
}
