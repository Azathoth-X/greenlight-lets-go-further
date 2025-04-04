package main

import (
	"errors"
	"expvar"
	"fmt"
	"greenlight/internal/data"
	"greenlight/internal/validator"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanicMiddleware(next http.HandlerFunc) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {

			if err := recover(); err != nil {
				w.Header().Set("Connection", "Close")

				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)

	})
}

func (app *application) rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {

	type client struct {
		lastSeen    time.Time
		rateLimiter *rate.Limiter
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.Host)

			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{rateLimiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].rateLimiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()

		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticateMiddleware(next http.HandlerFunc) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 && headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]
		v := validator.NewValidator()

		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.contextSetUser(r, user)

		next(w, r)
	})
}

func (app *application) requireAuthenticatedUserMiddleware(next http.HandlerFunc) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next(w, r)

	})
}

func (app *application) requireActivatedUserMiddleware(next http.HandlerFunc) http.HandlerFunc {

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := app.contextGetUser(r)

		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next(w, r)

	})

	return app.requireAuthenticatedUserMiddleware(fn)
}

func (app *application) requirePermissionsMiddleware(code string, next http.HandlerFunc) http.HandlerFunc {

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := app.contextGetUser(r)

		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		if !permissions.Includes(code) {
			app.notPermittedResponse(w, r)
			return
		}

		next(w, r)
	})

	return app.requireActivatedUserMiddleware(fn)
}

func (app *application) enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Vary", "Origin")

		origin := r.Header.Get("Origin")

		for i := range app.config.cors.trustedOrigins {
			if origin == app.config.cors.trustedOrigins[i] {

				w.Header().Set("Access-Control-Allow-Origin", origin)

				if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {

					w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
					w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

					w.WriteHeader(http.StatusOK)
					return
				}
				break
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) metricsMiddleware(next http.HandlerFunc) http.HandlerFunc {

	var (
		totalRequestReceived = expvar.NewInt("total_requests_received")
		totalResponseSent    = expvar.NewInt("total_response_sent")
		totalProcessingTime  = expvar.NewInt("total_processing_time_us")
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		curTime := time.Now()

		totalRequestReceived.Add(1)

		next.ServeHTTP(w, r)

		totalResponseSent.Add(1)

		duration := time.Since(curTime).Microseconds()
		totalProcessingTime.Add(duration)

	})

}
