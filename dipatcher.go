package awses

import (
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
)

type Dispatcher struct {
	Configs []*Config
	Next    httpserver.Handler

	handlers map[*Config]*Handler
}

func NewDispatcher(configs []*Config, next httpserver.Handler) *Dispatcher {
	rootSession := session.New()
	handlers := make(map[*Config]*Handler)
	for _, config := range configs {
		handlers[config] = NewHandler(config, rootSession)
	}

	return &Dispatcher{
		Configs: configs,
		Next:    next,

		handlers: handlers,
	}
}

func (d Dispatcher) ConfigForRequest(r *http.Request) *Config {
	for _, config := range d.Configs {
		if r.URL.Path == config.Path || strings.HasPrefix(r.URL.Path, config.Path+"/") {
			return config
		}
	}

	return nil
}

func (d Dispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	config := d.ConfigForRequest(r)
	if config == nil {
		return d.Next.ServeHTTP(w, r)
	}

	// strip the prefix before dispatching the handler
	r.URL.Path = strings.TrimPrefix(r.URL.Path, config.Path)
	return d.handlers[config].ServeHTTP(w, r)
}
