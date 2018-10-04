package api

import (
	"net/http"

	"github.com/vikramsk/rediproxy/pkg/cache"
)

const (
	apiPathCache = "/cache"
	paramKey     = "key"
)

// ProxyHandler is a wrapper for the
// proxy caching service.
type ProxyHandler struct {
	proxyService cache.Getter
}

// ensure that the handler implements
// the http.Handler interface
var _ = http.Handler(&ProxyHandler{})

// NewProxyHandler initializes a new ProxyHandler.
// It accepts the proxy service as a parameter.
func NewProxyHandler(ps cache.Getter) *ProxyHandler {
	return &ProxyHandler{
		proxyService: ps,
	}
}

// ServeHTTP implements the http handler for the proxy.
func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET" && r.URL.Path == apiPathCache:
		ph.handleGetRequest(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (ph *ProxyHandler) handleGetRequest(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get(paramKey)
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	val, err := ph.proxyService.Get(key)
	if err == cache.ErrKeyNotFound {
		w.WriteHeader(http.StatusNoContent)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(val))
}
