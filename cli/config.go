package cli

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	usage = `Usage: aws-signing [options...] <path>

Uses AWS environment variables when aws request signing is enabled.

 -d, --data <data>            HTTP POST data
                              Specify @- for stdin.

 -H, --header                 Pass custom header(s) to server
                              Defaults:
                                Content-Type: application/json

 -X, --request <command>      Specify request command to use
                              Default: GET

 -r, --reverse-proxy          Run reverse proxy using configuration
                              to reach aws endpoint.

 -p, --reverse-proxy-port     Configure reverse proxy server port.
                              Default: 9200
` + AwsConfig{}.HelpOptions()

	defaultReverseProxyPort = 9200
	logger                  = log.New(os.Stderr, "", 0)
)

type Config struct {
	Data             string
	Method           string
	Headers          http.Header
	ReverseProxy     bool
	ReverseProxyPort int
	Debug            bool
	Aws              AwsConfig
	RequestUrl       *url.URL
}

func Parse(args []string) (Config, error) {
	flags := flag.NewFlagSet("aws-signing", flag.ContinueOnError)
	flags.Usage = func() {
		logger.Println(usage)
	}

	var data string
	flags.StringVar(&data, "data", "", "data")
	flags.StringVar(&data, "d", "", "data (shorthand)")

	var method string
	flags.StringVar(&method, "request", "", "request method")
	flags.StringVar(&method, "X", "", "request method (shorthand)")

	var reverseProxy bool
	flags.BoolVar(&reverseProxy, "reverse-proxy", false, "run reverse proxy")
	flags.BoolVar(&reverseProxy, "r", false, "run reverse proxy (shorthand)")

	var reverseProxyPort int
	flags.IntVar(&reverseProxyPort, "reverse-proxy-port", defaultReverseProxyPort, "reverse proxy port")
	flags.IntVar(&reverseProxyPort, "p", defaultReverseProxyPort, "reverse proxy port (shorthand)")

	ac := &AwsConfig{}
	ac.AddFlags(flags)

	var debug bool
	flags.BoolVar(&debug, "debug", false, "debug")

	header := headerFlags{Headers: http.Header{}}
	flags.Var(&header, "H", "request header")
	flags.Var(&header, "header", "request header")

	if err := flags.Parse(args[1:]); err != nil {
		return Config{}, err
	}

	ac.Defaults()
	if err := ac.Normalize(); err != nil {
		return Config{}, err
	}
	ac.Dump()

	if ct := header.Headers.Get("Content-Type"); ct == "" {
		header.Headers.Set("Content-Type", "application/json")
	}

	remaining := flags.Args()
	if len(remaining) != 1 {
		return Config{}, fmt.Errorf(usage)
	}

	additional := strings.TrimPrefix(strings.TrimPrefix(remaining[0], "//"), "/")
	requestUrl, err := url.Parse(strings.TrimSuffix(ac.EndpointUrl.String(), "/") + "/" + additional)
	if err != nil {
		return Config{}, fmt.Errorf("error creating request url: %s", err)
	}

	c := Config{
		Data:             data,
		Method:           method,
		RequestUrl:       requestUrl,
		Headers:          header.Headers,
		ReverseProxy:     reverseProxy,
		ReverseProxyPort: reverseProxyPort,
		Debug:            debug,
		Aws:              *ac,
	}

	return c, nil
}

func (c Config) RequestBody() io.Reader {
	if c.Data == "@-" {
		return os.Stdin
	} else if c.Data == "" {
		return nil
	}
	return bytes.NewBufferString(c.Data)
}

func (c Config) Transport() (http.RoundTripper, error) {
	if c.Aws.Use {
		return c.Aws.Transport()
	}
	return http.DefaultTransport, nil
}
