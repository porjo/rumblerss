package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	shutdownTimeout = 10 * time.Second
)

var (
	maxTextLength int
	maxItemCount  int
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	port := flag.Int("port", 8080, "listen on this port")
	debug := flag.Bool("debug", false, "debug log output")
	flag.IntVar(&maxTextLength, "maxTextLength", 0, "limit each field to maximum number of characters (zero is unlimited)")
	flag.IntVar(&maxItemCount, "maxItemCount", 0, "limit the maximum number of feed items returned (zero is unlimited)")
	flag.Parse()

	var programLevel = new(slog.LevelVar) // Info by default
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))

	if *debug {
		programLevel.Set(slog.LevelDebug)
	}

	e := echo.New()
	e.GET("/", FeedHandler)

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		c.Logger().Errorf("error %s, request %q", err, c.Request().RequestURI)
		c.Echo().DefaultHTTPErrorHandler(err, c)
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		ReadTimeout:  httpServerReadTimeout,
		WriteTimeout: httpServerWriteTimeout,
		Handler:      e,
	}

	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			e.Logger.Errorf("error listening and serving: %s\n", err)
		}
	}()

	<-ctx.Done()
	log.Printf("Shutting down...")
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		e.Logger.Errorf("error shutting down http server: %s\n", err)
	}

	return nil
}
