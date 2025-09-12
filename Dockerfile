FROM alpine:latest

ARG UID=3000
ARG GID=3000

ENV LANGUAGE="en_US.UTF-8" \
    LANG="en_US.UTF-8"

RUN apk update && apk upgrade && \
    apk add --no-cache \
        bash \
        ca-certificates \
        git \
        openssh-client \
        gnupg \
        dumb-init \
        gopass && \
    addgroup -g ${GID} abc && \
    adduser -u ${UID} -G abc -h /home/abc -D abc && \
    mkdir /data && chown abc:abc /data && \
    rm -rf /tmp/* /var/tmp/*

COPY --chown=root:root --chmod=0755 entrypoint.sh /
COPY --chown=root:root --chmod=0755 external-secrets-gopass-webhook /

USER ${UID}:${GID}
WORKDIR /home/abc
VOLUME [ "/data" ]
ENV PASSWORD_STORE_DIR=/data

ENTRYPOINT ["/usr/bin/dumb-init", "/entrypoint.sh"]
