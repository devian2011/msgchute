package route

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swaggo/http-swagger/v2"

	"github.com/devian2011/msgchute/internal/io/web"
	"github.com/devian2011/msgchute/internal/io/web/endpoint/admin"
	"github.com/devian2011/msgchute/internal/io/web/endpoint/public"
	"github.com/devian2011/msgchute/internal/registry"
)

func RegisterRoutes(s *web.Server, handlers *registry.Handlers, m *registry.Middlewares) {
	r := mux.NewRouter()

	r.Use(appJsonMiddleware)
	r.Use(corsMiddleware)
	r.Use(m.Auth.Middleware)

	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}).Methods(http.MethodOptions)

	if s.WithSwagger() {
		r.PathPrefix("/api/swagger/").Handler(httpSwagger.Handler(
			httpSwagger.URL("/api/swagger/doc.json"),
		))
	}

	r = managementAPI(r, handlers)
	r = publicAPI(r, handlers)

	s.SetHandler(r)
}

func managementAPI(r *mux.Router, handlers *registry.Handlers) *mux.Router {
	r.Handle("/api/admin/v1/template", admin.NewTemplateFinderEndpoint(handlers.Admin.TemplateFinder)).
		Methods(http.MethodGet)
	r.Handle("/api/admin/v1/template", admin.NewTemplateCreationEndpoint(handlers.Admin.TemplateCreator)).
		Methods(http.MethodPost)
	r.Handle("/api/admin/v1/template/{code}", admin.NewTemplateUpdateEndpoint(handlers.Admin.TemplateUpdater)).
		Methods(http.MethodPut)

	r.Handle("/api/admin/v1/message", admin.NewMessageFinderEndpoint(handlers.Admin.MessageFinder)).
		Methods(http.MethodGet)
	r.Handle("/api/admin/v1/message/{id}", admin.NewMessageFinderByIDEndpoint(handlers.Admin.MessageFindByID)).
		Methods(http.MethodGet)

	r.Handle("/api/admin/v1/workers/status", admin.NewWorkerStatusEndpoint(handlers.Admin.WorkerStatus)).
		Methods(http.MethodGet)

	return r
}

func publicAPI(r *mux.Router, handlers *registry.Handlers) *mux.Router {
	r.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("pong"))
	})

	r.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("pong"))
	})

	r.Handle("/api/v1/send", public.NewSenderEndpoint(handlers.Public.Sender))
	r.Handle("/api/v1/preview", public.NewMessagePreviewEndpoint(handlers.Public.Preview))

	return r
}
