FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -o /out/agent ./cmd/agent


FROM alpine:3.20
RUN apk add --no-cache ca-certificates

COPY --from=build /out/agent /usr/local/bin/agent

ENTRYPOINT ["/usr/local/bin/agent"]
