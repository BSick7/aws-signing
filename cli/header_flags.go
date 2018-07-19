package cli

import (
	"fmt"
	"net/http"
	"strings"
)

type HeaderFlags struct {
	Headers http.Header
}

func (f *HeaderFlags) String() string {
	return "request headers"
}

func (f *HeaderFlags) Set(value string) error {
	tokens := strings.SplitN(value, ":", 2)
	if len(tokens) != 2 {
		return fmt.Errorf("invalid header %q", value)
	}

	f.Headers.Add(tokens[0], tokens[1])
	return nil
}
