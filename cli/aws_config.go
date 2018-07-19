package cli

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/BSick7/aws-signing/signing"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

var (
	defaultAwsEndpoint = "http://localhost:9200"
	defaultAwsService  = "es"
)

type AwsConfig struct {
	Use         bool
	Service     string
	endpoint    string
	EndpointUrl *url.URL
	dumpCreds   bool
}

func (c *AwsConfig) AddFlags(flags *flag.FlagSet) {
	flags.BoolVar(&c.Use, "aws", false, "use aws request signing")
	flags.BoolVar(&c.Use, "a", false, "use aws request signing (shorthand)")

	flags.StringVar(&c.endpoint, "aws-endpoint", "", "aws endpoint url")
	flags.StringVar(&c.endpoint, "e", "", "aws endpoint url (shorthand)")

	flags.StringVar(&c.Service, "aws-service", "", "aws service")
	flags.StringVar(&c.Service, "s", "", "aws service (shorthand)")

	flags.BoolVar(&c.dumpCreds, "creds", false, "emit creds")
}

func (AwsConfig) HelpOptions() string {
	return `
 -a, --aws                    Use AWS Request Signing
                              Default: false
                              Env Var: AWS_SIGNING

 -e, --aws-endpoint <url>     AWS Endpoint URL.
                              Default: http://localhost:9200
                              Env Var: AWS_ENDPOINT

 -s, --aws-service <service>  AWS Service.
                              Default: es
                              Env Var: AWS_SERVICE
`
}

func (c *AwsConfig) Defaults() {
	if _, ok := os.LookupEnv("AWS_SIGNING"); ok {
		c.Use = true
	}

	if env, ok := os.LookupEnv("AWS_ENDPOINT"); c.endpoint == "" && ok {
		c.endpoint = env
	}
	if c.endpoint == "" {
		c.endpoint = defaultAwsEndpoint
	}

	if svc, ok := os.LookupEnv("AWS_SERVICE"); svc == "" && ok {
		c.Service = svc
	}
	if c.Service == "" {
		c.Service = defaultAwsService
	}
}

func (c *AwsConfig) Normalize() error {
	eu, err := url.Parse(c.endpoint)
	if err != nil {
		return fmt.Errorf("error parsing endpoint url %q: %s", c.EndpointUrl, err)
	}
	c.EndpointUrl = eu
	return nil
}

func (c AwsConfig) Dump() {
	if c.Use && c.dumpCreds {
		cfg, _ := external.LoadDefaultAWSConfig()
		logger.Println(cfg.Credentials.Retrieve())
	}
}

func (c AwsConfig) Transport() (http.RoundTripper, error) {
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
