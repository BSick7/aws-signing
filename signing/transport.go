package signing

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type MissingSignerError struct{}

func (MissingSignerError) Error() string { return "signer is required to perform http request" }

type MissingServiceError struct{}

func (MissingServiceError) Error() string { return "aws service is required to perform http request" }

type MissingRegionError struct{}

func (MissingRegionError) Error() string { return "aws region is required to perform http request" }

// Signer represents an interface that v1 and v2 aws sdk follows to sign http requests
type Signer interface {
	Sign(r *http.Request, body io.ReadSeeker, service, region string, signTime time.Time) (http.Header, error)
}

// Creates a new transport that can be used by http.Client
// If region is unspecified, AWS_REGION environment variable is used
func NewTransport(signer Signer, service, region string) http.RoundTripper {
	return &Transport{
		signer:  signer,
		service: service,
		region:  region,
	}
}

// Transport implements http.RoundTripper and optionally wraps another RoundTripper
type Transport struct {
	BaseTransport http.RoundTripper
	signer        Signer
	service       string
	region        string
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.signer == nil {
		return nil, MissingSignerError{}
	}
	if t.service == "" {
		return nil, MissingServiceError{}
	}
	if t.region == "" {
		if t.region = os.Getenv("AWS_REGION"); t.region == "" {
			return nil, MissingRegionError{}
		}
	}

	baseTransport := t.BaseTransport
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}
	if h, ok := req.Header["Authorization"]; ok && len(h) > 0 && strings.HasPrefix(h[0], "AWS4") {
		return baseTransport.RoundTrip(req)
	}

	if err := t.sign(req); err != nil {
		return nil, fmt.Errorf("error signing request: %s", err)
	}
	return baseTransport.RoundTrip(req)
}

func (t *Transport) sign(req *http.Request) error {
	req.URL.Scheme = "https"
	if strings.Contains(req.URL.RawPath, "%2C") {
		req.URL.RawPath = escapePath(req.URL.RawPath, false)
	}

	// AWS forbids signed requests that contain X-Forwarded-For header
	req.Header.Del("X-Forwarded-For")

	date := time.Now()
	req.Header.Set("Date", date.Format(time.RFC3339))

	if body, err := t.rebuildBody(req); err != nil {
		return err
	} else if _, err := t.signer.Sign(req, body, t.service, t.region, date); err != nil {
		return fmt.Errorf("error signing request: %s", err)
	}
	return nil
}

func (t *Transport) rebuildBody(req *http.Request) (io.ReadSeeker, error) {
	if req.Body == nil {
		return nil, nil
	}

	d, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading http body to sign: %s", err)
	}
	req.Body = ioutil.NopCloser(bytes.NewReader(d))
	return bytes.NewReader(d), nil
}
