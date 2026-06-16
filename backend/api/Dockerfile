FROM golang:1.25-alpine AS build
WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend ./

RUN CGO_ENABLED=0 go build -o /out/api ./api


FROM alpine:3.20
RUN apk add --no-cache ca-certificates

COPY --from=build /out/api /usr/local/bin/api

EXPOSE 9090
ENTRYPOINT ["/usr/local/bin/api"]
