package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/BSick7/aws-signing/cli"
	"github.com/BSick7/aws-signing/config"
)

var (
	usage = `aws-curl [options...] <path>
Requests http service similar to curl with AWS signing.

Options:
 
 -d, --data <data>            HTTP POST data
                              Specify @- for stdin.

 -H, --header                 Pass custom header(s) to server
                              Defaults:
                                Content-Type: application/json

 -X, --request <command>      Specify request command to use
                              Default: GET
`
)

func main() {
	cfg, err := parse(os.Args)
	if err != nil {
		log.Fatalf("error parsing")
	}

	transport, err := cfg.Aws.Transport()
	if err != nil {
		log.Fatalf("error creating transport: %s\n", err)
	}
	client := &http.Client{Transport: transport}

	req, err := http.NewRequest(cfg.Method, cfg.RequestUrl(), cfg.RequestBody())
	if err != nil {
		log.Fatalf("error creating http request: %s", err)
	}
	req.Header = cfg.Headers

	log.Printf("%s %s\n", req.Method, req.URL)
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("response error: %s", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
		if _, err := io.Copy(os.Stdout, res.Body); err != nil {
			log.Fatalf("error writing response body: %s", err)
		}
	}
}

func parse(args []string) (config.Curl, error) {
	flags := flag.NewFlagSet("aws-curl", flag.ContinueOnError)
	flags.Usage = func() {
		log.Println(usage)
	}

	var data string
	flags.StringVar(&data, "data", "", "data")
	flags.StringVar(&data, "d", "", "data (shorthand)")

	var method string
	flags.StringVar(&method, "request", "", "request method")
	flags.StringVar(&method, "X", "", "request method (shorthand)")

	var debug bool
	flags.BoolVar(&debug, "debug", false, "debug")

	headers := cli.HeaderFlags{Headers: http.Header{}}
	flags.Var(&headers, "H", "request header")
	flags.Var(&headers, "header", "request header")

	aws := &cli.AwsArgs{}
	aws.AddFlags(flags)

	if err := flags.Parse(args[1:]); err != nil {
		return config.Curl{}, err
	}

	remaining := flags.Args()
	if len(remaining) != 1 {
		return config.Curl{}, fmt.Errorf(usage)
	}
	path := remaining[0]

	caws, err := aws.Config()
	if err != nil {
		return config.Curl{}, err
	}

	cl := config.Curl{
		Data:    data,
		Method:  method,
		Headers: headers.Headers,
		Path:    path,
		Aws:     caws,
		Debug:   debug,
	}

	result := config.MergeCurl(
		config.DefaultCurl,
		config.EnvCurl,
		cl,
	)

	return result, nil
}
