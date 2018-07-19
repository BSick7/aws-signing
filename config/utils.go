package config

import (
	"net/url"
	"os"
)

func hasEnvVar(v string) bool {
	_, has := os.LookupEnv(v)
	return has
}

func parseUrl(uri string, fallback *url.URL) *url.URL {
	if uri == "" {
		return nil
	}
	eu, _ := url.Parse(uri)
	return eu
}
