package config

import (
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	ApiAuthPass       string        `env:"API_AUTH_PASS"`
	ApiAuthUser       string        `env:"API_AUTH_USER"`
	ApiGetSecretPath  string        `env:"API_GET_PATH" envDefault:"/api/get"`
	ApiPostSecretPath string        `env:"API_POST_PATH" envDefault:"/api/post"`
	GitPushEnabled    bool          `env:"GIT_PUSH_ENABLED" envDefault:"false"`
	GitWebhookPath    string        `env:"GIT_WEBHOOK_PATH" envDefault:"/git"`
	GitWebhookSecret  string        `env:"GIT_WEBHOOK_SECRET,notEmpty"`
	GitWebhookType    string        `env:"GIT_WEBHOOK_TYPE,notEmpty"`
	ListenAddress     string        `env:"LISTEN_ADDRESS" envDefault:"0.0.0.0:3000"`
	RefreshInterval   time.Duration `env:"REFERSH_INTERVAL" envDefault:"1h"`
	RefreshLimit      time.Duration `env:"REFRESH_LIMIT" envDefault:"5m"`
}

func Init() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("error reading configuration from environment: %v", err)
	}
	if cfg.RefreshInterval > 0 && cfg.RefreshInterval < cfg.RefreshLimit {
		log.Errorln("auto-refresh interval cannot be shorter than the refresh limit")
		os.Exit(1)
	}

	if (cfg.ApiAuthUser == "") != (cfg.ApiAuthPass == "") {
		log.Errorln("auth requires both username and password")
		os.Exit(1)
	}
	return cfg
}
