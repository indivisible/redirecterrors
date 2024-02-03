package redirecterrors_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/indivisible/redirecterrors"
)

func TestBadConfig(t *testing.T) {
	cfg := redirecterrors.CreateConfig()
	cfg.Status = []string{}
	cfg.Target = ""

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	_, err := redirecterrors.New(ctx, next, cfg, "redirecterrors-plugin")
	if !assert(t, err != nil) {
		return
	}
	assert(t, err.Error() == "target url must be set")
}

// TODO: more tests: config parsing & non-intercepted response
func TestRedirect(t *testing.T) {
	cfg := redirecterrors.CreateConfig()
	cfg.Status = []string{"401", "402"}
	cfg.Target = "http://target/?status={status}&url={url}"

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) { rw.WriteHeader(401) })

	handler, err := redirecterrors.New(ctx, next, cfg, "redirecterrors-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	resp := recorder.Result()
	assertHeader(t, resp, "Location", "http://target/?status=401&url=http%3A%2F%2Flocalhost")
	assertCode(t, resp, 302)
}

func TestNoRedirect(t *testing.T) {
	cfg := redirecterrors.CreateConfig()
	cfg.Status = []string{}
	cfg.Target = "http://target/?status={status}&url={url}"

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := redirecterrors.New(ctx, next, cfg, "redirecterrors-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	resp := recorder.Result()
	assertCode(t, resp, 200)
	assertHeader(t, resp, "Location", "")
}

func assertCode(t *testing.T, resp *http.Response, expected int) {
	t.Helper()

	if resp.StatusCode != expected {
		t.Errorf("invalid status value: %d", resp.StatusCode)
	}
}

func assert(t *testing.T, condition bool) bool {
	t.Helper()

	if !condition {
		t.Error("Assertation failed")
	}
	return condition
}

func assertHeader(t *testing.T, resp *http.Response, key, expected string) {
	t.Helper()

	if resp.Header.Get(key) != expected {
		t.Errorf("invalid header value: %s", resp.Header.Get(key))
	}
}
