// Package redirecterrors traefik plugin to do external redirect on HTTP errors
package redirecterrors

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Config the plugin configuration.
type Config struct {
	Status       []string `json:"status,omitempty"`
	Target       string   `json:"target,omitempty"`
	OutputStatus int      `json:"outputStatus,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Status:       []string{},
		Target:       "",
		OutputStatus: 302,
	}
}

// RedirectErrors a RedirectErrors plugin.
type RedirectErrors struct {
	name           string
	next           http.Handler
	httpCodeRanges HTTPCodeRanges
	target         string
	outputStatus   int
}

// New creates a new RedirectErrors plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.Target) == 0 {
		return nil, fmt.Errorf("target url must be set")
	}

	httpCodeRanges, err := NewHTTPCodeRanges(config.Status)
	if err != nil {
		return nil, err
	}

	return &RedirectErrors{
		httpCodeRanges: httpCodeRanges,
		next:           next,
		name:           name,
		target:         config.Target,
		outputStatus:   config.OutputStatus,
	}, nil
}

func (a *RedirectErrors) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	catcher := newCodeCatcher(rw, a.httpCodeRanges)

	a.next.ServeHTTP(catcher, req)
	if !catcher.isFilteredCode() {
		return
	}
	code := catcher.getCode()
	println("Caught HTTP status code", code, "redirecting")

	// try to cobble together the original URL
	proto := req.Header.Get("X-Forwarded-Proto")
	host := req.Header.Get("X-Forwarded-Host")
	fullURL := req.URL.String()
	if len(proto) != 0 && len(host) != 0 {
		fullURL = proto + "://" + host
		fullURL += req.URL.RequestURI()
	} else {
		println("Missing proxy headers!")
	}

	location := a.target
	location = strings.ReplaceAll(location, "{status}", strconv.Itoa(code))
	location = strings.ReplaceAll(location, "{url}", url.QueryEscape(fullURL))

	println("New location:", location)
	rw.Header().Set("Location", location)
	rw.WriteHeader(a.outputStatus)
	_, err := io.WriteString(rw, "Redirecting")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
