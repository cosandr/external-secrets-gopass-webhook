# External Secrets webhook for gopass

Use gopass with ESO.

### Cloning private repos

You will need to mount `known_hosts` and the private key in `/home/abc/.ssh` if using SSH.

If using HTTPS, check git provider docs, typically the username/pass are included in the repo URL.

`REPO_URL="https://username:token@github.com/username/repo.git"`

### Env vars

Required:
- REPO_URL
- GIT_AUTHOR_NAME
- GIT_AUTHOR_EMAIL

One of:
- GPG_SECRET_FILE
- GPG_SECRET

### Running locally

Example `.env`

```sh
LOG_LEVEL=debug
GIT_WEBHOOK_TYPE=gitlab
GIT_WEBHOOK_SECRET="<token>"
GPG_SECRET_FILE=/gopass.gpg
GIT_AUTHOR_NAME=test
GIT_AUTHOR_EMAIL=test@example.com
REPO_URL=""
```

```sh
go build
docker build -t test .
docker run --rm -v $PWD/external-secrets-gopass-webhook.gpg:/gopass.gpg --env-file .env -p 3000:3000 test
```
