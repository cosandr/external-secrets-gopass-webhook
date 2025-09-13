package main

import (
	"context"
	"os"
	"time"

	"net/http"

	"github.com/cosandr/external-secrets-gopass-webhook/internal/api"
	"github.com/cosandr/external-secrets-gopass-webhook/internal/config"
	"github.com/cosandr/external-secrets-gopass-webhook/internal/git"
	"github.com/cosandr/external-secrets-gopass-webhook/internal/gopass"
	"github.com/cosandr/external-secrets-gopass-webhook/internal/logging"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/go-playground/webhooks/v6/gitlab"
	log "github.com/sirupsen/logrus"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	logging.Init()

	log.Infof("starting external-secrets-gopass-webhook version: %s (%s)", version, commit)

	config := config.Init()

	gopass, err := gopass.NewGopass(config.RefreshLimit)
	if err != nil {
		log.Fatalf("failed to initialize gopass: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := gopass.Close(ctx); err != nil {
			log.Errorf("failed to close gopass: %v", err)
		}
		log.Debugln("gopass closed")
	}()

	if config.RefreshInterval > 0 {
		log.Infof("auto-refreshing repo every %v", config.RefreshInterval)
		ticker := time.NewTicker(config.RefreshInterval)
		quit := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					log.Debugln("starting auto-refresh")
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					if err := gopass.Pull(ctx, false); err != nil {
						log.Errorf("auto-refresh failed: %v", err)
					}
					log.Debugln("auto-refresh completed")
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()
	}

	switch config.GitWebhookType {
	case "github":
		hook, _ := github.New(github.Options.Secret(config.GitWebhookSecret))
		http.HandleFunc("POST "+config.GitWebhookPath, func(w http.ResponseWriter, r *http.Request) {
			git.HandleGithubWebhook(hook, gopass, r)
		})
		log.Info("listening for Gitlab webhooks")
	case "gitlab":
		hook, _ := gitlab.New(gitlab.Options.Secret(config.GitWebhookSecret))
		http.HandleFunc("POST "+config.GitWebhookPath, func(w http.ResponseWriter, r *http.Request) {
			git.HandleGitlabWebhook(hook, gopass, r)
		})
		log.Info("listening for Gitlab webhooks")
	default:
		log.Errorf("unknown webhook type: %s", config.GitWebhookType)
		os.Exit(1)
	}

	http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	http.HandleFunc("GET "+config.ApiGetSecretPath, func(w http.ResponseWriter, r *http.Request) {
		if config.ApiAuthUser != "" {
			username, password, ok := r.BasicAuth()
			if !ok || (username != config.ApiAuthUser) || (password != config.ApiAuthPass) {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		api.HandleGetSecret(gopass, w, r)
	})

	http.HandleFunc("POST "+config.ApiPostSecretPath, func(w http.ResponseWriter, r *http.Request) {
		if config.ApiAuthUser != "" {
			username, password, ok := r.BasicAuth()
			if !ok || (username != config.ApiAuthUser) || (password != config.ApiAuthPass) {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		api.HandlePostSecret(gopass, config.GitPushEnabled, w, r)
	})

	http.ListenAndServe(config.ListenAddress, nil)
}
