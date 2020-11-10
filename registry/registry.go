package registry

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type LogfCallback func(format string, args ...interface{})

/*
 * Discard log messages silently.
 */
func Quiet(format string, args ...interface{}) {
	/* discard logs */
}

/*
 * Pass log messages along to Go's "log" module.
 */
func Log(format string, args ...interface{}) {
	log.Printf(format, args...)
}

type OptionFunc func(r *Registry)

func SetLogf(logf LogfCallback) OptionFunc {
	return func(r *Registry) {
		r.logf = logf
	}
}

func SetHttpClient(client *http.Client) OptionFunc {
	return func(r *Registry) {
		r.client = client
	}
}

type Registry struct {
	url    string
	client *http.Client
	logf   LogfCallback
}

type Option struct {
	Logf LogfCallback
}

/*
 * Create a new Registry with the given URL and credentials, then Ping()s it
 * before returning it to verify that the registry is available.
 *
 * You can, alternately, construct a Registry manually by populating the fields.
 * This passes http.DefaultTransport to WrapTransport when creating the
 * http.Client.
 */
func New(registryURL, username, password string, options ...OptionFunc) (*Registry, error) {
	transport := http.DefaultTransport

	return newFromTransport(registryURL, username, password, transport, options...)
}

/*
 * Create a new Registry, as with New, using an http.Transport that disables
 * SSL certificate verification.
 */
func NewInsecure(registryURL, username, password string, options ...OptionFunc) (*Registry, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			// TODO: Why?
			InsecureSkipVerify: true, //nolint:gosec
		},
	}

	return newFromTransport(registryURL, username, password, transport, options...)
}

/*
 * Given an existing http.RoundTripper such as http.DefaultTransport, build the
 * transport stack necessary to authenticate to the Docker registry API. This
 * adds in support for OAuth bearer tokens and HTTP Basic auth, and sets up
 * error handling this library relies on.
 */
func WrapTransport(transport http.RoundTripper, url, username, password string) http.RoundTripper {
	tokenTransport := &TokenTransport{
		Transport: transport,
		Username:  username,
		Password:  password,
	}
	basicAuthTransport := &BasicTransport{
		Transport: tokenTransport,
		URL:       url,
		Username:  username,
		Password:  password,
	}
	errorTransport := &ErrorTransport{
		Transport: basicAuthTransport,
	}
	return errorTransport
}

func newFromTransport(registryURL, username, password string, transport http.RoundTripper, options ...OptionFunc) (*Registry, error) {
	url := strings.TrimSuffix(registryURL, "/")
	transport = WrapTransport(transport, url, username, password)
	registry := &Registry{
		url: url,
		client: &http.Client{
			Transport: transport,
		},
		logf: Log,
	}

	for _, optionFunc := range options {
		optionFunc(registry)
	}

	if err := registry.Ping(); err != nil {
		return nil, err
	}

	return registry, nil
}

func (r *Registry) generateUrl(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", r.url, pathSuffix)
	return url
}

func (r *Registry) Ping() error {
	url := r.generateUrl("/v2/")
	r.logf("registry.ping url=%s", url)
	resp, err := r.client.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	return err
}
