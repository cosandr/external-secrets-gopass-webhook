#!/bin/bash

set -eo pipefail

if [ -z "$REPO_URL" ] || [ -z "$GIT_AUTHOR_NAME" ] || [ -z "$GIT_AUTHOR_EMAIL" ]; then
    echo "REPO_URL,GIT_AUTHOR_NAME and GIT_AUTHOR_EMAIL are required"
    exit 1
fi

if [ -n "$GPG_SECRET_FILE" ]; then
    if [ -f "$GPG_SECRET_FILE" ]; then
        echo "Importing GPG key from file"
        gpg -q --import "$GPG_SECRET_FILE"
    else
        echo "$GPG_SECRET_FILE not found"
        exit 1
    fi
elif [ -n "$GPG_SECRET" ]; then
    echo "Importing GPG key from env"
    echo "$GPG_SECRET" | base64 -d | gpg -q --import
else
    echo "GPG_SECRET or GPG_SECRET_FILE is required"
    exit 1
fi

# Clone with git since gopass prints secrets to stdout
mkdir -p "$PASSWORD_STORE_DIR"
# Try to fetch first, in case repo was already cloned
if git -C "$PASSWORD_STORE_DIR" fetch -p &>/dev/null; then
    git -C "$PASSWORD_STORE_DIR" reset --hard origin/HEAD &>/dev/null
    echo "Updated exiting repo"
else
    git clone "$REPO_URL" "$PASSWORD_STORE_DIR"
fi
# Sanity check, assign to variable to fail on error
num_secrets="$(gopass ls -f | wc -l)"
echo "Initialized gopass, found $num_secrets secrets"

exec /external-secrets-gopass-webhook
