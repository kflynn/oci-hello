package main

import (
	"context"
	"net/http"
	"os"
	// "time"

	"github.com/sirupsen/logrus"

	// "github.com/datawire/dlib/dcontext"
	// "github.com/datawire/dlib/derror"
	// "github.com/datawire/dlib/dexec"
	"github.com/datawire/dlib/dgroup"
	"github.com/datawire/dlib/dhttp"
	"github.com/datawire/dlib/dlog"
	// "github.com/datawire/dlib/dtime"
)

// This is an example main() program entry-point that shows how all the pieces of dlib can fit
// together and complement each other.
func main() {
	// Start with the background Context as the root Context.
	ctx := context.Background()

	// The default backend for dlog is pretty good, but for the sake of example, let's customize
	// it a bit.
	ctx = dlog.WithLogger(ctx, func() dlog.Logger {
		// Let's have the backend be logrus.  The default backend is already logrus, but
		// ours will be customized.
		logrusLogger := logrus.New()
		// The dlog default is InfoLevel; let's crank it up to DebugLevel.
		logrusLogger.Level = logrus.DebugLevel
		// Now turn that in to a dlog.Logger backend, so we can pass it to dlog.WithLogger.
		return dlog.WrapLogrus(logrusLogger)
	}())

	// We're going to be doing several tasks in parallel, so we'll use "dgroup" to manage our
	// group of goroutines.
	grp := dgroup.NewGroup(ctx, dgroup.GroupConfig{
		// Enable signal handling for graceful shutdown.  The user can stop the program by
		// sending it SIGINT with Ctrl-C, and that will start a graceful shutdown.  If that
		// graceful shutdown takes too long, and the user hits Ctrl-C again, then it will
		// start a not-so-graceful shutdown.
		//
		// This shutdown will be signaled to the worker goroutines through the Context that
		// gets passed to them.  The mechanism by which the Context signals both graceful
		// and not-so-graceful shutdown is what "dcontext" is for.
		EnableSignalHandling: true,
	})

	// One of those tasks will be running an HTTP server.
	grp.Go("http", func(ctx context.Context) error {
		// We'll be using a *dhttp.ServerConfig instead of an *http.Server, but it works
		// very similarly to *http.Server, everything else in the stdlib net/http package is
		// still valid; we'll still be using plain-old http.ResponseWriter and *http.Request
		// and http.HandlerFunc.
		dlog.Infof(ctx, "Starting webserver on :8080")
		cfg := &dhttp.ServerConfig{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				dlog.Debugln(r.Context(), "handling HTTP request")
				_, _ = w.Write([]byte("Hello, world!\n"))
			}),
		}
		// ListenAndServe will gracefully shut down according to ctx; we don't need to worry
		// about separately calling .Shutdown() or .Close() like we would for *http.Server
		// (those methods don't even exist on dhttp.ServerConfig).  During a graceful
		// shutdown, it will stop listening and close idle connections, but will wait on any
		// active connections; during a not-so-graceful shutdown it will forcefully close
		// any active connections.
		//
		// If the server itself needs to log anything, it will use dlog according to ctx.
		// The Request.Context() passed to the Handler function will inherit from ctx, and
		// so the Handler will also log according to ctx.
		//
		// And, on the end-user-facing side of things, this supports HTTP/2, where
		// *http.Server.ListenAndServe wouldn't.
		return cfg.ListenAndServe(ctx, ":8080")
	})

	if err := grp.Wait(); err != nil {
		dlog.Errorf(ctx, "finished with error: %v", err)
		os.Exit(1)
	}
}
