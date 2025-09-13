package gopass

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"github.com/gopasspw/gopass/pkg/gopass/api"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
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
	mu             sync.Mutex
}

func NewGopass(refreshLimit time.Duration) (*Gopass, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
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
	// Can only run one git operation at once
	g.mu.Lock()
	defer g.mu.Unlock()
	log.Debugln("starting gopass git repo refresh")
	// gp.Sync() is not implemented
	// Use git directly to handle force pushes and also because gopass sync attempts to push
	pullOK := true
	cmd := exec.CommandContext(ctx, "gopass", "git", "pull", "--prune", "--rebase")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		pullOK = false
		log.Warnf("git pull failed '%v': %s", err, stderr.String())
	}
	// Reset if pull failed
	if !pullOK {
		cmd := exec.CommandContext(ctx, "gopass", "git", "fetch", "--prune")
		stderr.Reset()
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
		log.Debugln("git reset OK")
	} else {
		log.Debugln("git pull OK")
	}
	log.Infoln("gopass git repo refreshed")
	return nil
}

func (g *Gopass) Push(ctx context.Context, name string, value string) error {
	// Pull first to ensure we don't overwrite stuff
	g.Pull(ctx, true)
	log.Debugf("starting pushing '%s'", name)
	sec := secrets.New()
	sec.SetPassword(value)
	if err := g.gp.Set(ctx, name, sec); err != nil {
		return err
	}
	log.Infof("pushed secret '%s'", name)
	return nil
}

func (g *Gopass) Close(ctx context.Context) error {
	return g.gp.Close(ctx)
}
