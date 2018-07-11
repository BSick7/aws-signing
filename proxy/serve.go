package proxy

import (
	"fmt"
	"log"
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
	log.Printf("listening on %s\n", server.Addr)
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
		Director: func(request *http.Request) {
			template.Director(request)
			request.Host = request.URL.Hostname()
		},
		// TODO: We need to rewrite aws_signing_client to drop X-Forwarded-For
		//         or wrap this transport around a round tripper that drops it
		// AWS forbids X-Forwarded-For header on signed requests
		Transport: httpClient.Transport,
	}
	return rp, nil
}
