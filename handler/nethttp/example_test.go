package wasm

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	httpwasm "github.com/http-wasm/http-wasm-host-go"
	"github.com/http-wasm/http-wasm-host-go/internal/test"
)

var (
	responseBody = "{\"hello\": \"world\"}"

	serveJson = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		w.Write([]byte(responseBody)) // nolint
	})

	servePath = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "text/plain")
		w.Write([]byte(r.URL.Path)) // nolint
	})
)

func Example_auth() {
	ctx := context.Background()

	// Configure and compile the WebAssembly guest binary. In this case, it is
	// an auth interceptor.
	mw, err := NewMiddleware(ctx, test.AuthWasm)
	if err != nil {
		log.Panicln(err)
	}
	defer mw.Close(ctx)

	// Create the real request handler.
	next := serveJson

	// Wrap this with an interceptor implemented in WebAssembly.
	wrapped := mw.NewHandler(ctx, next)

	// Start the server with the wrapped handler.
	ts := httptest.NewServer(wrapped)
	defer ts.Close()

	// Invoke some requests, only one of which should pass
	headers := []http.Header{
		{"NotAuthorization": {"1"}},
		{"Authorization": {""}},
		{"Authorization": {"Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ=="}},
		{"Authorization": {"0"}},
	}

	for _, header := range headers {
		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		if err != nil {
			log.Panicln(err)
		}
		req.Header = header

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Panicln(err)
		}
		resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			fmt.Println("OK")
		case http.StatusUnauthorized:
			fmt.Println("Unauthorized")
		default:
			log.Panicln("unexpected status code", resp.StatusCode)
		}
	}

	// Output:
	// Unauthorized
	// Unauthorized
	// OK
	// Unauthorized
}

func Example_log() {
	ctx := context.Background()
	logger := func(_ context.Context, message string) { fmt.Println(message) }

	// Configure and compile the WebAssembly guest binary. In this case, it is
	// a logging interceptor.
	mw, err := NewMiddleware(ctx, test.LogWasm, httpwasm.Logger(logger))
	if err != nil {
		log.Panicln(err)
	}
	defer mw.Close(ctx)

	// Create the real request handler.
	next := serveJson

	// Wrap this with an interceptor implemented in WebAssembly.
	wrapped := mw.NewHandler(ctx, next)

	// Start the server with the wrapped handler.
	ts := httptest.NewServer(wrapped)
	defer ts.Close()

	// Make a client request and print the contents to the same logger
	resp, err := ts.Client().Get(ts.URL)
	if err != nil {
		log.Panicln(err)
	}
	defer resp.Body.Close()

	// Ensure the response body was still readable!
	body, _ := io.ReadAll(resp.Body)
	if want, have := responseBody, string(body); want != have {
		log.Panicf("unexpected response body, want: %q, have: %q", want, have)
	}

	// Output:
	// request body:
	// response body:
	// {"hello": "world"}
}

func Example_router() {
	ctx := context.Background()

	// Configure and compile the WebAssembly guest binary. In this case, it is
	// an example request router.
	mw, err := NewMiddleware(ctx, test.RouterWasm)
	if err != nil {
		log.Panicln(err)
	}
	defer mw.Close(ctx)

	// Wrap the real handler with an interceptor implemented in WebAssembly.
	wrapped := mw.NewHandler(ctx, servePath)

	// Start the server with the wrapped handler.
	ts := httptest.NewServer(wrapped)
	defer ts.Close()

	// Invoke some requests, only one of which should pass
	paths := []string{
		"",
		"nothosst",
		"host/a",
	}

	for _, p := range paths {
		url := fmt.Sprintf("%s/%s", ts.URL, p)
		resp, err := ts.Client().Get(url)
		if err != nil {
			log.Panicln(err)
		}
		defer resp.Body.Close()
		content, _ := io.ReadAll(resp.Body)
		fmt.Println(string(content))
	}

	// Output:
	// hello world
	// hello world
	// /a
}

func Example_redact() {
	ctx := context.Background()

	// Configure and compile the WebAssembly guest binary. In this case, it is
	// an example response redact.
	secret := "open sesame"
	mw, err := NewMiddleware(ctx, test.RedactWasm,
		httpwasm.GuestConfig([]byte(secret)))
	if err != nil {
		log.Panicln(err)
	}
	defer mw.Close(ctx)

	var body string
	serveBody := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "text/plain")
		w.Write([]byte(body)) // nolint
	})

	// Wrap the real handler with an interceptor implemented in WebAssembly.
	wrapped := mw.NewHandler(ctx, serveBody)

	// Start the server with the wrapped handler.
	ts := httptest.NewServer(wrapped)
	defer ts.Close()

	bodies := []string{
		secret,
		"hello world",
		fmt.Sprintf("hello %s world", secret),
	}

	for _, b := range bodies {
		body = b

		resp, err := ts.Client().Get(ts.URL)
		if err != nil {
			log.Panicln(err)
		}
		defer resp.Body.Close()
		content, _ := io.ReadAll(resp.Body)
		fmt.Println(string(content))
	}

	// Output:
	// ###########
	// hello world
	// hello ########### world
}
