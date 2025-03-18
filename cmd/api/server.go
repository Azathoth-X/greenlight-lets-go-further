package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	app.logger.Info("Server starting", "addr", srv.Addr, "env", app.config.env)

	errShutdown := make(chan error)
	go func() {

		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

		s := <-quit

		app.logger.Info("server shutting down ", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			errShutdown <- err
		}

		app.logger.Info("completing bg tasks")

		app.wg.Wait()

		errShutdown <- nil
	}()

	err := srv.ListenAndServe()

	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	err = <-errShutdown
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}
