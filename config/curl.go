package config

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	DefaultCurl = Curl{
		Method: "GET",
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Path: "/",
		Aws:  DefaultsAws,
	}
	EnvCurl = Curl{
		Aws: EnvAws,
	}
)

type Curl struct {
	Data    string      `hcl:"-"`
	Method  string      `hcl:"-"`
	Headers http.Header `hcl:"-"`
	Path    string      `hcl:"-"`
	Aws     Aws         `hcl:"aws"`
	Debug   bool        `hcl:"-"`
}

func MergeCurl(cfgs ...Curl) Curl {
	rv := Curl{Headers: http.Header{}}
	for _, cur := range cfgs {
		if cur.Data != "" {
			rv.Data = cur.Data
		}
		if cur.Method != "" {
			rv.Method = cur.Method
		}
		if cur.Headers != nil {
			for k, v := range cur.Headers {
				rv.Headers[k] = v
			}
		}
		if cur.Path != "" {
			rv.Path = cur.Path
		}
		rv.Aws = MergeAws(rv.Aws, cur.Aws)
		rv.Debug = rv.Debug || cur.Debug
	}
	return rv
}

func (c Curl) RequestUrl() string {
	additional := strings.TrimPrefix(strings.TrimPrefix(c.Path, "//"), "/")
	requestUrl, _ := url.Parse(strings.TrimSuffix(c.Aws.EndpointUrl().String(), "/") + "/" + additional)
	return requestUrl.String()
}

func (c Curl) RequestBody() io.Reader {
	if c.Data == "@-" {
		return os.Stdin
	} else if c.Data == "" {
		return nil
	}
	return bytes.NewBufferString(c.Data)
}
