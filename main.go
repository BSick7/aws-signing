package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/BSick7/aws-signing/cli"
)

var logger = log.New(os.Stderr,"",0)

func main() {
	config, err := cli.Parse(os.Args)
	if err != nil {
		logger.Println(err)
	}

	req, err := http.NewRequest(config.Method, config.RequestUrl.String(), config.RequestBody())
	if err != nil {
		logger.Fatalf("error creating http request: %s", err)
	}

	client, err := config.HttpClient()
	if err != nil {
		logger.Fatalf("error creating http client: %s", err)
	}

	log.Printf("%s %s\n", req.Method, req.URL)
	res, err := client.Do(req)
	if err != nil {
		logger.Fatalf("response error: %s", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
		if _, err := io.Copy(os.Stdout, res.Body); err != nil {
			logger.Fatalf("error writing response body: %s", err)
		}
	}
}
