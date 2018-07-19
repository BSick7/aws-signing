package config

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/BSick7/aws-signing/signing"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

var (
	DefaultsAws = Aws{
		Use:         false,
		EndpointUrl: &url.URL{Scheme: "http", Host: "localhost:9200"},
		Service:     "es",
	}
	EnvAws = Aws{
		Use:         hasEnvVar("AWS_SIGNING"),
		EndpointUrl: parseUrl(os.Getenv("AWS_ENDPOINT"), nil),
		Service:     os.Getenv("AWS_SERVICE"),
	}
)

type Aws struct {
	Use         bool     `hcl:"enabled"`
	Service     string   `hcl:"service"`
	EndpointUrl *url.URL `hcl:"endpoint-url"`
}

func MergeAws(cfgs ...Aws) Aws {
	rv := Aws{}
	for _, cur := range cfgs {
		rv.Use = rv.Use || cur.Use
		if cur.EndpointUrl != nil {
			rv.EndpointUrl = cur.EndpointUrl
		}
		if cur.Service != "" {
			rv.Service = cur.Service
		}
	}
	return rv
}

func (c Aws) Transport() (http.RoundTripper, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading aws config: %s", err)
	}
	if region := os.Getenv("AWS_REGION"); region != "" {
		cfg.Region = region
	}
	signer := v4.NewSigner(cfg.Credentials)
	return signing.NewTransport(signer, c.Service, cfg.Region), nil
}
