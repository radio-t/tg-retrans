FROM golang:1.23-alpine as build

ARG GIT_BRANCH
ARG GITHUB_SHA

ADD . /build
WORKDIR /build

RUN version=${GIT_BRANCH}-${GITHUB_SHA:0:7}-$(date +%Y%m%dT%H:%M:%S)
RUN echo "version=$version"
RUN go build -o /build/tg-retrans -ldflags "-X main.revision=${version} -s -w"


#FROM alpine:3.20
FROM umputun/baseimage:app-latest

# enables automatic changelog generation by tools like Dependabot
LABEL org.opencontainers.image.source="https://github.com/radio-t/tg-retrans"

RUN \
    set -xe; \
    echo "**** install runtime ****" && \
    apk add --update --no-cache ffmpeg && \
    rm -rf /var/cache/apk/* && \
    echo "**** quick test ffmpeg ****" && \
    ldd /usr/bin/ffmpeg && \
    /usr/bin/ffmpeg -version

COPY --from=build /build/tg-retrans /srv/tg-retrans
COPY /logo-dark.png /srv
RUN chown -R app:app /srv

WORKDIR /srv
USER app:app

ENTRYPOINT ["/srv/tg-retrans"]
