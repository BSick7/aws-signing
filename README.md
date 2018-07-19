# aws-signing

This is both a library and CLI designed to aid AWS request signing.

The library provides a golang developer with several abilities stemming from a `RoundTripper`.
This library provides constructs on top of the `RoundTripper` to aid other http functions.

## Library

The transport can be created with [aws-sdk-go](https://github.com/aws/aws-sdk-go) or [aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2).
This transport can then be used with an http client.

```go
import (
	"net/http"
	
	"github.com/aws/aws-sdk-go-v2/aws/credentials"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/BSick7/aws-signing/signing"
)

var credsProvider aws.CredentialsProvider
// ... set credentials ...
signer := v4.NewSigner(credsProvider)
transport := signing.NewTransport(signer, "es", "us-east-1")
httpClient := &http.Client{
	Transport: transport,
}
``` 

## Install CLI

```
$ go get -u github.com/BSick7/aws-signing
```

## CLI Usage

This is a CLI designed to be used with AWS request signing.
The usage is a very stripped-down version of curl.

This tool provides env vars to enable concise commands.

```
$ ./aws-signing
Usage: aws-signing [options...] <path>

 -a, --aws                Use AWS Request Signing
                          Default: false
                          Env Var: ES_AWS

 -d, --data <data>        HTTP POST data
                          Specify @- for stdin.

 -e, --endpoint <url>     Elasticsearch endpoint url.
                          Default: http://localhost:9200
                          Env Var: ES_ENDPOINT

 -X, --request <command>  Specify request command to use
                          Default: GET
```

## CLI AWS Request Signing

If aws request signing is enabled, this tool uses the same chain that the aws cli uses.
This allows you to work seamlessly between aws cli tool without setting up additional configuration for access keys.

## Credit

This library was inspired by [https://github.com/sha1sum/aws_signing_client](https://github.com/sha1sum/aws_signing_client).
This library uses better configuration without modifying `http.DefaultClient`.
Also, logging is not configured on a global level.
Additionally, a reverse proxy construct is added for use in trusted environments.
