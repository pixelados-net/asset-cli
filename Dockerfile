FROM golang:1.26-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates
COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY platform ./platform

ARG VERSION=0.0.1
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/asset-cli ./cmd

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/asset-cli /asset-cli
ENTRYPOINT ["/asset-cli"]
