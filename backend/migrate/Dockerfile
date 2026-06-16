FROM golang:1.25-alpine AS build
WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend ./

RUN CGO_ENABLED=0 go build -o /out/migrate ./migrate


FROM alpine:3.20
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=build /out/migrate /usr/local/bin/migrate
COPY backend/migrations ./migrations

ENTRYPOINT ["/usr/local/bin/migrate"]
