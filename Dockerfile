FROM alpine:3.20

RUN \
  set -xe; \
  echo "**** install runtime ****" && \
    apk add --update --no-cache ffmpeg nushell && \
    rm -rf /var/cache/apk/* && \
  echo "**** quick test ffmpeg ****" && \
    ldd /usr/bin/ffmpeg && \
    /usr/bin/ffmpeg -version

COPY /entrypoint.nu /
COPY /logo-dark.png /

SHELL ["/usr/bin/nu", "-c"]

ENTRYPOINT ["/entrypoint.nu"]
