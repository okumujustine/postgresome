FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate


FROM alpine:3.20
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=build /out/migrate /usr/local/bin/migrate
COPY migrations ./migrations

ENTRYPOINT ["/usr/local/bin/migrate"]
