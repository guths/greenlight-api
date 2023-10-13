package main

import (
	"context"
	"net/http"

	"github.com/guths/greenlight-api/internal/data"
)

type contextKey string

// can be pass a string directly to a context but is best practice to use a custom type to avoid colition with own or third part codes
const userContextKey = contextKey("user")

func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)

	if !ok {
		panic("missing user value in request context")
	}

	return user
}
