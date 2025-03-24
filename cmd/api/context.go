package main

import (
	"context"
	"greenlight/internal/data"
	"net/http"
)

type userContext string

const userContextKey = userContext("user")

func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing values in user context")
	}

	return user
}
