package signing

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewReverseProxy(endpointUrl *url.URL, transport http.RoundTripper) *httputil.ReverseProxy {
	template := httputil.NewSingleHostReverseProxy(endpointUrl)
	if transport == nil {
		return template
	}

	return &httputil.ReverseProxy{
		Director: func(request *http.Request) {
			template.Director(request)
			request.Host = request.URL.Hostname()
		},
		Transport: transport,
	}
}
