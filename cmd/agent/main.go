// Command agent runs the KOSMOS collector agent.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	cpucontracts "github.com/natsubrei/kosmos/contracts/cpu"
	"github.com/natsubrei/kosmos/internal/agent/bus"
	cpudomain "github.com/natsubrei/kosmos/internal/agent/domain/cpu"
	"github.com/natsubrei/kosmos/internal/agent/report"
	"github.com/natsubrei/kosmos/internal/agent/runtime"
)

func main() {
	interval := flag.Duration("interval", time.Second, "collection tick interval")
	flag.Parse()

	if err := run(*interval); err != nil {
		fmt.Fprintln(os.Stderr, "agent:", err)
		os.Exit(1)
	}
}

func run(interval time.Duration) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	b := bus.New()
	b.Subscribe(cpucontracts.Domain, report.NewJSONLines(os.Stdout))

	bundle, err := cpudomain.New(interval)
	if err != nil {
		return err
	}
	a := runtime.New(b)
	if err := a.Register(bundle); err != nil {
		return err
	}
	return a.Run(ctx)
}
