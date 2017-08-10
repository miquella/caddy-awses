package awses

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
)

type Handler struct {
	Config *Config

	manager *ElasticsearchManager
}

func NewHandler(config *Config, rootSession *session.Session) *Handler {
	return &Handler{
		Config: config,

		manager: NewElasticsearchManager(rootSession, config.Role),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	// parse the path
	path := r.URL.Path
	if path == "" {
		path = "/"
	}

	region := h.Config.Region
	if region == "" {
		region, path = splitNextComponent(path)
	}

	domain := h.Config.Domain
	if domain == "" {
		domain, path = splitNextComponent(path)
	}

	// render the corresponding response
	if region == "" {
		return h.renderMissingRegion(w)
	} else if domain == "" {
		return h.renderMissingDomain(w, region)
	} else {
		return h.proxyRequest(w, r, region, domain, path)
	}
}

func (h *Handler) renderMissingRegion(w http.ResponseWriter) (int, error) {
	http.Error(w, "An AWS region must be provided", http.StatusBadRequest)
	return 0, nil
}

func (h *Handler) renderMissingDomain(w http.ResponseWriter, region string) (int, error) {
	domains, err := h.manager.ListDomains(region)
	if err != nil {
		http.Error(w, "An AWS ES domain name must be provided", http.StatusInternalServerError)
		return 0, nil
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusBadRequest)

	fmt.Fprint(w, "An AWS ES domain name must be provided. Available domain names:\n\n")
	for _, domain := range domains {
		fmt.Fprintf(w, "%s\n", domain)
	}
	return 0, nil
}

func (h *Handler) proxyRequest(w http.ResponseWriter, r *http.Request, region, domain, path string) (int, error) {
	reverseProxy, err := h.manager.GetProxy(region, domain)
	if err != nil {
		if err == ErrDomainNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err == ErrInvalidDomainName {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusBadGateway)
		}
		return 0, err
	}

	r.URL.Path = path
	reverseProxy.ServeHTTP(w, r)
	return 0, nil
}

func splitNextComponent(path string) (string, string) {
	split := strings.SplitN(strings.TrimLeft(path, "/"), "/", 2)
	if len(split) == 2 {
		return split[0], "/" + split[1]
	} else {
		return split[0], "/"
	}
}
