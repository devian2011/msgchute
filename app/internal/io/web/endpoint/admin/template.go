package admin

import (
	"net/http"

	"github.com/devian2011/msgchute/internal/dto"
)

// Template finder
type TemplateFinderHandler interface {
	Handle(filter *dto.MessageTemplateFilter) ([]*dto.Template, uint64, error)
}

type TemplateFinderEndpoint struct {
	h TemplateFinderHandler
}

func NewTemplateFinderEndpoint(h TemplateFinderHandler) *TemplateFinderEndpoint {
	return &TemplateFinderEndpoint{h: h}
}

func (e *TemplateFinderEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

// Template creation

type TemplateCreationHandler interface {
	Handle(template *dto.Template) (*dto.Template, error)
}
type TemplateCreationEndpoint struct {
	h TemplateCreationHandler
}

func NewTemplateCreationEndpoint(h TemplateCreationHandler) *TemplateCreationEndpoint {
	return &TemplateCreationEndpoint{h: h}
}

func (e *TemplateCreationEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

// Template Update

type TemplateUpdateHandler interface {
	Handle(template *dto.Template) (*dto.Template, error)
}
type TemplateUpdateEndpoint struct {
	h TemplateUpdateHandler
}

func NewTemplateUpdateEndpoint(h TemplateUpdateHandler) *TemplateUpdateEndpoint {
	return &TemplateUpdateEndpoint{h: h}
}

func (e *TemplateUpdateEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
