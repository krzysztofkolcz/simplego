package daemon

import (
	"net/http"
)

type ServeMux struct {
	httpServeMux http.ServeMux
	BaseURL      string
}

func NewServeMux(baseURL string) *ServeMux {
	return &ServeMux{
		httpServeMux: http.ServeMux{},
		BaseURL:      baseURL,
	}
}

func (m *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.httpServeMux.ServeHTTP(w, r)
}

func (m *ServeMux) HandleFunc(
	pattern string,
	handler func(http.ResponseWriter, *http.Request),
) {
	m.httpServeMux.HandleFunc(pattern, handler)
}
