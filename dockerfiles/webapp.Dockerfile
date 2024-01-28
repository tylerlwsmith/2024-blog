FROM golang:1.21 AS base

RUN mkdir -p /srv/app

WORKDIR /srv/app

FROM base AS development

RUN go install github.com/cosmtrek/air@latest

CMD ["air"]

FROM base AS build

COPY webapp/go.mod webapp/go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o webapp

FROM scratch AS production

COPY --from=build --chown=root:root /srv/app/webapp /

CMD ["/webapp"]
