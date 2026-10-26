// Command server receives Records over HTTP JSON and lists them.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	cpucontracts "github.com/natsubrei/kosmos/contracts/cpu"
	"github.com/natsubrei/kosmos/internal/bus"
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

func run(listen string) error {
	b := bus.New()
	st := store.NewMemory()
	b.Subscribe(cpucontracts.Domain, st)
	fmt.Fprintf(os.Stderr, "server: listening on %s\n", listen)
	return http.ListenAndServe(listen, httpapi.Handler(b, st))
}
