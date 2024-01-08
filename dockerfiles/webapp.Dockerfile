FROM golang:1.21

RUN mkdir -p /srv/app

WORKDIR /srv/app

RUN go install github.com/cosmtrek/air@latest

# Pre-copy/cache go.mod for pre-downloading dependencies and only redownloading
# them in subsequent builds if they change.
COPY webapp/go.mod webapp/go.sum ./
RUN go mod download && go mod verify

# COPY . .
# RUN go build -v -o /usr/local/bin/app ./...

CMD ["air"]
