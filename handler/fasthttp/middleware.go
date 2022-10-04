package wasm

import (
	"context"
	"strconv"

	"github.com/valyala/fasthttp"

	httpwasm "github.com/http-wasm/http-wasm-host-go"
	"github.com/http-wasm/http-wasm-host-go/api/handler"
	internalhandler "github.com/http-wasm/http-wasm-host-go/internal/handler"
)

type Middleware handler.Middleware[fasthttp.RequestHandler]

// compile-time check to ensure middleware implements Middleware.
var _ Middleware = &middleware{}

type middleware struct {
	runtime *internalhandler.Runtime
	// TODO: pool
	guest *internalhandler.Guest
}

func NewMiddleware(ctx context.Context, guest []byte, options ...httpwasm.Option) (Middleware, error) {
	r, err := internalhandler.NewRuntime(ctx, guest, &host{}, options...)
	if err != nil {
		return nil, err
	}
	g, err := r.NewGuest(ctx)
	if err != nil {
		return nil, err
	}
	return &middleware{runtime: r, guest: g}, nil
}

type host struct{}

// GetPath implements the same method as documented on handler.Host.
func (h host) GetPath(ctx context.Context) string {
	r := &ctx.(*fasthttp.RequestCtx).Request
	return string(r.URI().Path())
}

// SetPath implements the same method as documented on handler.Host.
func (h host) SetPath(ctx context.Context, path string) {
	r := &ctx.(*fasthttp.RequestCtx).Request
	r.URI().SetPath(path)
}

// GetRequestHeader implements the same method as documented on
// handler.Host.
func (h host) GetRequestHeader(ctx context.Context, name string) (string, bool) {
	r := &ctx.(*fasthttp.RequestCtx).Request
	if value := r.Header.Peek(name); value == nil {
		return "", false
	} else {
		return string(value), true
	}
}

// Next implements the same method as documented on handler.Host.
func (h host) Next(ctx context.Context) {
	fastCtx := ctx.(*fasthttp.RequestCtx)
	fastCtx.UserValue("next").(fasthttp.RequestHandler)(fastCtx)
}

// SetResponseHeader implements the same method as documented on handler.Host.
func (h host) SetResponseHeader(ctx context.Context, name, value string) {
	r := &ctx.(*fasthttp.RequestCtx).Response
	r.Header.Set(name, value)
}

// SendResponse implements the same method as documented on handler.Host.
func (h host) SendResponse(ctx context.Context, statusCode uint32, body []byte) {
	r := &ctx.(*fasthttp.RequestCtx).Response
	if body != nil {
		r.Header.Set("Content-Length", strconv.Itoa(len(body)))
		r.AppendBody(body)
	}
	r.SetStatusCode(int(statusCode))
}

// NewHandler implements the same method as documented on handler.Middleware.
func (w *middleware) NewHandler(ctx context.Context, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return (&guest{handle: w.guest.Handle, next: next}).Handle
}

// Close implements the same method as documented on handler.Middleware.
func (w *middleware) Close(ctx context.Context) error {
	return w.runtime.Close(ctx)
}

type guest struct {
	handle func(ctx context.Context) (err error)
	next   fasthttp.RequestHandler
}

// Handle implements RequestHandler.Handle
func (w *guest) Handle(ctx *fasthttp.RequestCtx) {
	ctx.SetUserValue("next", w.next)
	if err := w.handle(ctx); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
}
