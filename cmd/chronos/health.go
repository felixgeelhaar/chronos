package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/felixgeelhaar/chronos/client"
	"github.com/felixgeelhaar/chronos/internal/config"
)

// defaultHealthTimeout bounds the probe. Container runtimes kill a
// HEALTHCHECK that overruns its own --timeout anyway; failing fast keeps
// the exit code meaningful rather than letting the runtime SIGKILL us.
const defaultHealthTimeout = 3 * time.Second

// runHealth probes GET /health on a running server and exits non-zero
// when it does not report healthy.
//
// It exists so the runtime image can declare a HEALTHCHECK. The image is
// built on gcr.io/distroless/static, which ships no shell and no curl,
// so the only executable available to a health probe is the chronos
// binary itself.
func runHealth(args []string) error {
	fs := flag.NewFlagSet("health", flag.ContinueOnError)
	addr := fs.String("addr", "", "base URL to probe (default http://127.0.0.1:$CHRONOS_HTTP_PORT)")
	timeout := fs.Duration("timeout", defaultHealthTimeout, "probe timeout")
	if err := fs.Parse(args); err != nil {
		return NewUserError("health: %v", err)
	}

	base := *addr
	if base == "" {
		// Probe loopback, never cfg.HTTPHost: containerised
		// deployments set CHRONOS_HTTP_HOST=0.0.0.0, which is a bind
		// wildcard and not a valid dial target. The port is the one
		// piece of the listener address a probe genuinely needs.
		base = fmt.Sprintf("http://127.0.0.1:%d", config.Default().HTTPPort)
	}

	c, err := client.New(base, client.WithTimeout(*timeout))
	if err != nil {
		return NewUserError("health: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// /health is deliberately exempt from BearerAuth (see
	// internal/api/middleware.go), so the probe needs no token even
	// when CHRONOS_API_TOKEN is set.
	if err := c.Health(ctx); err != nil {
		return NewSystemError(err, "health: %s is not healthy: %v", base, err)
	}

	fmt.Println("healthy")
	return nil
}
