package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/BSick7/aws-signing/cli"
)

func Serve(config cli.Config) error {
	rp, err := newReverseProxy(config)
	if err != nil {
		return fmt.Errorf("error creating reverse proxy: %s", err)
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.ReverseProxyPort),
		Handler: rp,
	}
	return server.ListenAndServe()
}

func newReverseProxy(config cli.Config) (*httputil.ReverseProxy, error) {
	template := httputil.NewSingleHostReverseProxy(config.EndpointUrl)
	if !config.UseAws {
		return template, nil
	}

	httpClient, err := config.HttpClient()
	if err != nil {
		return nil, err
	}

	rp := &httputil.ReverseProxy{
		Director:  template.Director,
		Transport: httpClient.Transport,
	}
	return rp, nil
}
