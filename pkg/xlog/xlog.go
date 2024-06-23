package xlog

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ContextHandler struct {
	slog.Handler
	keys []any
}

type (
	requestPathKey   struct{}
	requestMethodKey struct{}
)

func NewContextHandler(h slog.Handler, keys ...any) ContextHandler {
	return ContextHandler{Handler: h, keys: keys}
}

func NewDefaultContextHandler(h slog.Handler) ContextHandler {
	keys := []any{requestPathKey{}, requestMethodKey{}}
	return ContextHandler{Handler: h, keys: keys}
}

// Handler implements [slog.Handler].
func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.Handler == nil {
		return fmt.Errorf("xlog: handler is missing")
	}

	r.AddAttrs(h.observe(ctx)...)
	return h.Handler.Handle(ctx, r)
}

// WithAttrs implements [slog.Handler].
func (h ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.Handler == nil {
		return h
	}

	return ContextHandler{Handler: h.Handler.WithAttrs(attrs), keys: h.keys}
}

// WithGroup implements [slog.Handler].
func (h ContextHandler) WithGroup(name string) slog.Handler {
	if h.Handler == nil {
		return h
	}

	return ContextHandler{Handler: h.Handler.WithGroup(name), keys: h.keys}
}

func (h ContextHandler) observe(ctx context.Context) (as []slog.Attr) {
	for _, k := range h.keys {
		a, ok := ctx.Value(k).(slog.Attr)
		if !ok {
			continue
		}
		a.Value = a.Value.Resolve()
		as = append(as, a)
	}
	return
}

func NewEnhancedContext(ctx context.Context, r *http.Request) context.Context {
	ctx = context.WithValue(ctx, requestPathKey{}, slog.String("http.request.path", r.URL.Path))
	ctx = context.WithValue(ctx, requestMethodKey{}, slog.String("http.request.method", r.Method))
	return ctx
}

func EnhanceContext(c *gin.Context) {
	ctx := c.Request.Context()
	ctx = NewEnhancedContext(ctx, c.Request)
	c.Request = c.Request.Clone(ctx)
	c.Next()
}
