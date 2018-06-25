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

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/sha1sum/aws_signing_client"
)

var (
	usage = `Usage: aws-signing [options...] <path>

Uses AWS environment variables when aws request signing is enabled.

 -a, --aws                Use AWS Request Signing
                          Default: false
                          Env Var: ES_AWS

 -d, --data <data>        HTTP POST data
                          Specify @- for stdin.

 -e, --endpoint <url>     Elasticsearch endpoint url.
                          Default: http://localhost:9200
                          Env Var: ES_ENDPOINT

 -H, --header             Pass custom header(s) to server
                          Defaults:
                            Content-Type: application/json

 -X, --request <command>  Specify request command to use
                          Default: GET

`

	defaultEndpointUrl = "http://localhost:9200"
	logger             = log.New(os.Stderr, "", 0)
)

type Config struct {
	Data       string
	Method     string
	RequestUrl *url.URL
	Headers    http.Header
	UseAws     bool
}

func Parse(args []string) (Config, error) {
	flags := flag.NewFlagSet("aws-signing", flag.ContinueOnError)
	flags.Usage = func() {
		logger.Println(usage)
	}

	var data string
	flags.StringVar(&data, "data", "", "data")
	flags.StringVar(&data, "d", "", "data (shorthand)")

	var endpointUrl string
	flags.StringVar(&endpointUrl, "endpoint", "", "endpoint url")
	flags.StringVar(&endpointUrl, "e", "", "endpoint url (shorthand)")

	var method string
	flags.StringVar(&method, "request", "", "request method")
	flags.StringVar(&method, "X", "", "request method (shorthand)")

	var useAws bool
	flags.BoolVar(&useAws, "aws", false, "use aws request signing")
	flags.BoolVar(&useAws, "a", false, "use aws request signing (shorthand)")

	header := headerFlags{Headers: http.Header{}}
	flags.Var(&header, "H", "request header")
	flags.Var(&header, "header", "request header")

	if err := flags.Parse(args[1:]); err != nil {
		return Config{}, err
	}

	if _, ok := os.LookupEnv("ES_AWS"); ok {
		useAws = true
	}

	if env, ok := os.LookupEnv("ES_ENDPOINT"); endpointUrl == "" && ok {
		endpointUrl = env
	}
	if endpointUrl == "" {
		endpointUrl = defaultEndpointUrl
	}
	eu, err := url.Parse(endpointUrl)
	if err != nil {
		return Config{}, fmt.Errorf("error parsing endpoint url %q: %s", endpointUrl, err)
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

	return Config{
		Data:       data,
		Method:     method,
		RequestUrl: requestUrl,
		UseAws:     useAws,
		Headers:    header.Headers,
	}, nil
}

func (c Config) RequestBody() io.Reader {
	if c.Data == "@-" {
		return os.Stdin
	} else if c.Data == "" {
		return nil
	}
	return bytes.NewBufferString(c.Data)
}

func (c Config) HttpClient() (*http.Client, error) {
	if c.UseAws {
		cfg, err := external.LoadDefaultAWSConfig()
		if err != nil {
			return nil, err
		}
		if region := os.Getenv("AWS_REGION"); region != "" {
			cfg.Region = region
		}
		signer := v4.NewSigner(cfg.Credentials)
		aws_signing_client.New(signer, nil, "es", cfg.Region)
	}
	return http.DefaultClient, nil
}
