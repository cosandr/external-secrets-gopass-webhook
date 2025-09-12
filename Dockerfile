FROM debian:13-slim

ARG UID=3000
ARG GID=3000

ENV LANGUAGE="en_US.UTF-8" \
    LANG="en_US.UTF-8" \
    DEBIAN_FRONTEND=noninteractive

ADD --chown=root:root --chmod=0644 https://packages.gopass.pw/repos/gopass/gopass-archive-keyring.gpg /usr/share/keyrings/gopass-archive-keyring.gpg
# Need ca-certificates first, copy to /tmp and move later
COPY --chown=root:root --chmod=0644 gopass.sources /tmp/gopass.sources

RUN apt-get update && \
    apt-get upgrade -y && \
    echo '**** apt: install deps ****' && \
    apt-get install -y --no-install-recommends \
        ca-certificates locales git openssh-client gpg dumb-init && \
    echo '*** apt: install gopass' && \
    mv /tmp/gopass.sources /etc/apt/sources.list.d/gopass.sources && \
    apt-get update -y && apt-get install -y --no-install-recommends gopass gopass-archive-keyring  && \
    echo '**** user: Create abc user and group ****' && \
    groupadd --gid ${GID} abc && useradd --create-home --gid ${GID} --uid ${UID} abc && \
    mkdir /data && chown abc:abc /data && \
    echo '**** locale setup ****' && \
    locale-gen en_US.UTF-8 && \
    echo "**** cleanup ****" && \
    apt-get clean && \
    rm -rf \
        /tmp/* \
        /var/lib/apt/lists/* \
        /var/tmp/*

COPY --chown=root:root --chmod=0755 entrypoint.sh /
COPY --chown=root:root --chmod=0755 external-secrets-gopass-webhook /

USER ${UID}:${GID}
WORKDIR /home/abc
VOLUME [ "/data" ]
ENV PASSWORD_STORE_DIR=/data

ENTRYPOINT ["/usr/bin/dumb-init", "/entrypoint.sh"]
