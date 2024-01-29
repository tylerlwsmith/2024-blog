FROM golang:1.21 AS base

RUN mkdir -p /srv/app
WORKDIR /srv/app

COPY webapp/go.mod webapp/go.sum ./
RUN go mod download && go mod verify

FROM base AS development

RUN go install github.com/cosmtrek/air@latest
CMD ["air"]

FROM base AS build

COPY webapp/ .
# Build without anything that could possibly link outside.
# https://stackoverflow.com/a/55106860/7759523
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /srv/app/webapp

FROM scratch AS production

COPY --from=build --chown=root:root /srv/app/webapp /
CMD ["/webapp"]
