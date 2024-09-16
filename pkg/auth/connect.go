package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"connectrpc.com/connect"
)

type key int

const infoKey key = iota

func GetUserEmailConnect(ctx context.Context) (string, bool) {
	user := GetInfo(ctx)
	if userS, ok := user.(string); ok {
		return userS, true
	}

	return "", false
}

func MustGetUserEmailConnect(ctx context.Context) string {
	email, ok := GetUserEmailConnect(ctx)
	if !ok || email == "" {
		panic("empty user email")
	}

	return email
}

func WithInfo(ctx context.Context, info any) context.Context {
	if info == nil {
		return ctx
	}
	return context.WithValue(ctx, infoKey, info)
}

func GetInfo(ctx context.Context) any {
	return ctx.Value(infoKey)
}

func AuthenticationInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, err error) {
			auth := req.Header().Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(auth, prefix) {
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("no Authorization header"))
			}

			token := auth[len(prefix):]

			user, isValid := ValidateJWT(token)
			if !isValid {
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid token"))
			}

			ctx = WithInfo(ctx, user)

			resp, err = next(ctx, req)
			return
		}
	}
}

func CopyAuthHeader[T any](new *connect.Request[T], old *http.Request) error {
	if old.Header.Get("Authorization") != "" {
		new.Header().Add("Authorization", old.Header.Get("Authorization"))
	}

	cookie, err := old.Cookie(authCookieName)
	if err != nil {
		return fmt.Errorf("get cookie: %w", err)
	}

	val, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return fmt.Errorf("unescape cookie: %w", err)
	}

	new.Header().Add("Authorization", fmt.Sprintf("Bearer %s", val))
	return nil
}
