# aws-signing

This is a CLI designed to be used with AWS request signing.
The usage is a very stripped-down version of curl.

This tool provides env vars to enable concise commands.

## Install

```
$ go get -u github.com/BSick7/aws-signing
```

## Usage

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

## AWS Request Signing

If aws request signing is enabled, this tool uses the same chain that the aws cli uses.
This allows you to work seamlessly between aws cli tool without setting up additional configuration for access keys.
