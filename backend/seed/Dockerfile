FROM golang:1.25-alpine AS build
WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend ./

RUN CGO_ENABLED=0 go build -o /out/seed ./seed


FROM alpine:3.20
RUN apk add --no-cache ca-certificates

COPY --from=build /out/seed /usr/local/bin/seed

ENTRYPOINT ["/usr/local/bin/seed"]
