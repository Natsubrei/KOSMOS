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
	"github.com/natsubrei/kosmos/internal/agent/control"
	cpudomain "github.com/natsubrei/kosmos/internal/agent/domain/cpu"
	"github.com/natsubrei/kosmos/internal/agent/hostid"
	"github.com/natsubrei/kosmos/internal/agent/report"
	"github.com/natsubrei/kosmos/internal/agent/runtime"
	"github.com/natsubrei/kosmos/internal/agent/upload"
	"github.com/natsubrei/kosmos/internal/bus"
)

func main() {
	interval := flag.Duration("interval", time.Second, "collection tick interval")
	server := flag.String("server", "", "KOSMOS server base URL; empty skips upload")
	user := flag.String("user", "dev", "X-User-Id sent with uploads")
	host := flag.String("host", "", "machine identity stamped on Records; default is the kernel hostname")
	flag.Parse()

	if err := run(*interval, *server, *user, *host); err != nil {
		fmt.Fprintln(os.Stderr, "agent:", err)
		os.Exit(1)
	}
}

func run(interval time.Duration, server, user, hostOverride string) error {
	host, err := hostid.Resolve(hostOverride)
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	b := bus.New()
	b.Subscribe(cpucontracts.Domain, report.NewJSONLines(os.Stdout))
	if server != "" {
		u := upload.New(server, user)
		u.Start(ctx)
		b.Subscribe(cpucontracts.Domain, u)
	}

	bundle, err := cpudomain.New(interval)
	if err != nil {
		return err
	}
	a := runtime.New(b, host)
	if err := a.Register(bundle); err != nil {
		return err
	}
	if server != "" {
		ctrl := control.Client{
			BaseURL: server,
			User:    user,
			Host:    host,
			Apply:   a.Apply,
		}
		go func() { _ = ctrl.Run(ctx) }()
	} else if err := a.Apply([]string{cpucontracts.Domain}); err != nil {
		return err
	}
	return a.Run(ctx)
}
