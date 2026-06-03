FROM --platform=$BUILDPLATFORM gsoci.azurecr.io/giantswarm/golang:1.26.4-alpine3.22 AS builder

ARG TARGETOS TARGETARCH

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -v .

FROM gsoci.azurecr.io/giantswarm/alpine:3.23.4
COPY --from=builder /usr/src/app/frontmatter-validator /usr/local/bin/frontmatter-validator
RUN chmod +x /usr/local/bin/frontmatter-validator

WORKDIR /app

ENTRYPOINT ["/usr/local/bin/frontmatter-validator"]
