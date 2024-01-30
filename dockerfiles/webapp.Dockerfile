FROM golang:1.21 AS base

RUN mkdir -p /srv/app
WORKDIR /srv/app

COPY webapp/go.mod webapp/go.sum ./
RUN go mod download && go mod verify

FROM base AS development

# Ensure development works on Linux.
ARG USER_GID=1234
ARG USER_UID=1234
RUN if getent passwd $USER_UID; then userdel "$(getent passwd $USER_UID | cut -d: -f1)"; fi
RUN if getent group $USER_GID; then groupdel "$(getent group $USER_GID | cut -d: -f1)"; fi
RUN groupadd --gid $USER_GID app
RUN useradd --uid $USER_UID --gid $USER_GID --create-home app
RUN chown -R app:app /srv/app
RUN chown -R app:app /go
USER app

RUN go install github.com/cosmtrek/air@latest
CMD ["air"]

FROM base AS build

COPY webapp/ .
# Build without anything that could possibly link outside.
# https://stackoverflow.com/a/55106860/7759523
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /srv/app/webapp

# Setup /etc/passwd file for production scratch container.
# https://darrencodes.blog/2023/07/01/configuring-a-non-root-user-in-a-scratch-docker-image/
RUN echo "app:x:1234:1234::/:/bin/false" > /srv/scratch-passwd

FROM scratch AS production

WORKDIR /

COPY --from=build /srv/scratch-passwd /etc/passwd
COPY --from=build --chown=root:root /srv/app/webapp /

USER app
CMD ["/webapp"]
