package gopass

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"github.com/gopasspw/gopass/pkg/gopass/api"
)

// ErrSecretNotFound is a custom error type for when a secret cannot be found.
type ErrSecretNotFound struct {
	Name string
}

// Error makes ErrSecretNotFound satisfy the error interface.
func (e *ErrSecretNotFound) Error() string {
	return fmt.Sprintf("could not find secret '%s'", e.Name)
}

type Gopass struct {
	gp             *api.Gopass // https://pkg.go.dev/github.com/gopasspw/gopass@v1.15.16/pkg/gopass/api#example-package
	refreshLimiter *rate.Limiter
}

func NewGopass(ctx context.Context, refreshLimit time.Duration) (*Gopass, error) {
	gp, err := api.New(ctx)
	if err != nil {
		log.Fatalf("failed to initialize gopass: %v", err)
	}
	return &Gopass{
		gp:             gp,
		refreshLimiter: rate.NewLimiter(rate.Every(refreshLimit), 1),
	}, nil
}

func (g *Gopass) GetSecret(ctx context.Context, name string) (string, error) {
	ls, err := g.gp.List(ctx)
	if err != nil {
		return "", err
	}
	if !slices.Contains(ls, name) {
		return "", &ErrSecretNotFound{Name: name}
	}
	sec, err := g.gp.Get(ctx, name, "latest")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(sec.Bytes())), nil
}

func (g *Gopass) Pull(ctx context.Context, force bool) error {
	if !g.refreshLimiter.Allow() && !force {
		log.Debugln("refresh not allowed yet")
		return nil
	}
	log.Debugln("starting gopass git repo refresh")
	// gp.Sync() is not implemented
	// Use git directly to handle force pushes and also because gopass sync attempts to push
	cmd := exec.CommandContext(ctx, "gopass", "git", "fetch", "--prune")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch failed '%v': %s", err, stderr.String())
	}
	log.Debugln("git fetch OK")
	cmd = exec.CommandContext(ctx, "gopass", "git", "reset", "--hard", "origin/HEAD")
	stderr.Reset()
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git reset failed '%v': %s", err, stderr.String())
	}
	log.Infoln("gopass git repo refreshed")
	return nil
}

func (g *Gopass) Close(ctx context.Context) error {
	return g.gp.Close(ctx)
}
