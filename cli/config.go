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

	"github.com/BSick7/aws-signing/signing"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

var (
	usage = `Usage: aws-signing [options...] <path>

Uses AWS environment variables when aws request signing is enabled.

 -d, --data <data>        HTTP POST data
                          Specify @- for stdin.

 -a, --aws                Use AWS Request Signing
                          Default: false
                          Env Var: AWS_SIGNING

 -e, --endpoint <url>     AWS Endpoint URL.
                          Default: http://localhost:9200
                          Env Var: AWS_ENDPOINT

 -s, --service <service>  AWS Service.
                          Default: es
                          Env Var: AWS_SERVICE

 -H, --header             Pass custom header(s) to server
                          Defaults:
                            Content-Type: application/json

 -X, --request <command>  Specify request command to use
                          Default: GET

 -r, --reverse-proxy      Run reverse proxy using configuration
                          to reach aws endpoint.

 -p, --reverse-proxy-port Configure reverse proxy server port.
                          Default: 9200
`

	defaultEndpointUrl      = "http://localhost:9200"
	defaultAwsService       = "es"
	defaultReverseProxyPort = 9200
	logger                  = log.New(os.Stderr, "", 0)
)

type Config struct {
	Data             string
	Method           string
	EndpointUrl      *url.URL
	RequestUrl       *url.URL
	Headers          http.Header
	UseAws           bool
	AwsService       string
	ReverseProxy     bool
	ReverseProxyPort int
	Debug            bool
}

func Parse(args []string) (Config, error) {
	flags := flag.NewFlagSet("aws-signing", flag.ContinueOnError)
	flags.Usage = func() {
		logger.Println(usage)
	}

	var data string
	flags.StringVar(&data, "data", "", "data")
	flags.StringVar(&data, "d", "", "data (shorthand)")

	var useAws bool
	flags.BoolVar(&useAws, "aws", false, "use aws request signing")
	flags.BoolVar(&useAws, "a", false, "use aws request signing (shorthand)")

	var endpointUrl string
	flags.StringVar(&endpointUrl, "endpoint", "", "endpoint url")
	flags.StringVar(&endpointUrl, "e", "", "endpoint url (shorthand)")

	var awsService string
	flags.StringVar(&awsService, "service", "", "aws service")
	flags.StringVar(&awsService, "s", "", "aws service (shorthand)")

	var method string
	flags.StringVar(&method, "request", "", "request method")
	flags.StringVar(&method, "X", "", "request method (shorthand)")

	var reverseProxy bool
	flags.BoolVar(&reverseProxy, "reverse-proxy", false, "run reverse proxy")
	flags.BoolVar(&reverseProxy, "r", false, "run reverse proxy (shorthand)")

	var reverseProxyPort int
	flags.IntVar(&reverseProxyPort, "reverse-proxy-port", defaultReverseProxyPort, "reverse proxy port")
	flags.IntVar(&reverseProxyPort, "p", defaultReverseProxyPort, "reverse proxy port (shorthand)")

	var creds bool
	flags.BoolVar(&creds, "creds", false, "emit creds")

	var debug bool
	flags.BoolVar(&debug, "debug", false, "debug")

	header := headerFlags{Headers: http.Header{}}
	flags.Var(&header, "H", "request header")
	flags.Var(&header, "header", "request header")

	if err := flags.Parse(args[1:]); err != nil {
		return Config{}, err
	}

	if _, ok := os.LookupEnv("AWS_SIGNING"); ok {
		useAws = true
	}

	if env, ok := os.LookupEnv("AWS_ENDPOINT"); endpointUrl == "" && ok {
		endpointUrl = env
	}
	if endpointUrl == "" {
		endpointUrl = defaultEndpointUrl
	}
	eu, err := url.Parse(endpointUrl)
	if err != nil {
		return Config{}, fmt.Errorf("error parsing endpoint url %q: %s", endpointUrl, err)
	}

	if svc, ok := os.LookupEnv("AWS_SERVICE"); svc == "" && ok {
		awsService = svc
	}
	if awsService == "" {
		awsService = defaultAwsService
	}

	if ct := header.Headers.Get("Content-Type"); ct == "" {
		header.Headers.Set("Content-Type", "application/json")
	}

	remaining := flags.Args()
	if len(remaining) != 1 {
		return Config{}, fmt.Errorf(usage)
	}

	additional := strings.TrimPrefix(strings.TrimPrefix(remaining[0], "//"), "/")
	requestUrl, err := url.Parse(strings.TrimSuffix(eu.String(), "/") + "/" + additional)
	if err != nil {
		return Config{}, fmt.Errorf("error creating request url: %s", err)
	}

	c := Config{
		Data:             data,
		Method:           method,
		RequestUrl:       requestUrl,
		UseAws:           useAws,
		AwsService:       awsService,
		EndpointUrl:      eu,
		Headers:          header.Headers,
		ReverseProxy:     reverseProxy,
		ReverseProxyPort: reverseProxyPort,
		Debug:            debug,
	}

	if useAws && creds {
		cfg, _ := external.LoadDefaultAWSConfig()
		logger.Println(cfg.Credentials.Retrieve())
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
	if c.UseAws {
		cfg, err := external.LoadDefaultAWSConfig()
		if err != nil {
			return nil, fmt.Errorf("error loading aws config: %s", err)
		}
		if region := os.Getenv("AWS_REGION"); region != "" {
			cfg.Region = region
		}
		signer := v4.NewSigner(cfg.Credentials)
		return signing.NewTransport(signer, c.AwsService, cfg.Region), nil
	}
	return http.DefaultTransport, nil
}
