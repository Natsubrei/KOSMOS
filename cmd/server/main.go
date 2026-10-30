// Command server receives Records over HTTP JSON, runs Analyzers, and lists Records and Reports.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	cpucontracts "github.com/natsubrei/kosmos/contracts/cpu"
	"github.com/natsubrei/kosmos/internal/bus"
	cpuanalyzer "github.com/natsubrei/kosmos/internal/server/analyzer/cpu"
	"github.com/natsubrei/kosmos/internal/server/control"
	"github.com/natsubrei/kosmos/internal/server/diagnose"
	"github.com/natsubrei/kosmos/internal/server/httpapi"
	"github.com/natsubrei/kosmos/internal/server/store"
)

func main() {
	listen := flag.String("listen", ":8080", "HTTP listen address")
	flag.Parse()

	if err := run(*listen); err != nil {
		fmt.Fprintln(os.Stderr, "server:", err)
		os.Exit(1)
	}
}

func route(records, ctrl http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/control" {
			ctrl.ServeHTTP(w, r)
			return
		}
		records.ServeHTTP(w, r)
	})
}

func run(listen string) error {
	b := bus.New()
	st := store.NewMemory()
	b.Subscribe(cpucontracts.Domain, st)
	loop := diagnose.New(st, []diagnose.Analyzer{cpuanalyzer.New()})
	go loop.Run(context.Background())
	hub := control.NewHub(func(string) []string { return []string{cpucontracts.Domain} })
	fmt.Fprintf(os.Stderr, "server: listening on %s\n", listen)
	return http.ListenAndServe(listen, route(httpapi.Handler(b, st, loop), hub.Handler()))
}
