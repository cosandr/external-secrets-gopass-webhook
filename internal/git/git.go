package git

import (
	"context"
	"net/http"
	"time"

	"github.com/cosandr/external-secrets-gopass-webhook/internal/gopass"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/go-playground/webhooks/v6/gitlab"
	log "github.com/sirupsen/logrus"
)

func HandleGithubWebhook(hook *github.Webhook, gp *gopass.Gopass, r *http.Request) {
	payload, err := hook.Parse(r, github.PushEvent)
	if err != nil {
		if err == github.ErrEventNotFound || err == github.ErrEventNotSpecifiedToParse {
			log.Debugf("received non-push GitHub event")
		} else {
			log.Errorf("GitHub push event failure: %v", err)
		}
		return
	}

	log.Debugf("received GitHub push event: %v", payload)
	go triggerRefresh(gp)
}

func HandleGitlabWebhook(hook *gitlab.Webhook, gp *gopass.Gopass, r *http.Request) {
	payload, err := hook.Parse(r, gitlab.PushEvents)
	if err != nil {
		if err == gitlab.ErrEventNotFound || err == gitlab.ErrEventNotSpecifiedToParse {
			log.Debugf("received non-push Gitlab event")
		} else {
			log.Errorf("Gitlab push event failure: %v", err)
		}
		return
	}

	log.Debugf("received Gitlab push event: %v", payload)
	go triggerRefresh(gp)
}

func triggerRefresh(gp *gopass.Gopass) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	gp.Pull(ctx, true)
}
