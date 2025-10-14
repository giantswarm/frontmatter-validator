FROM gsoci.azurecr.io/giantswarm/golang:1.25.3-alpine3.22 AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v .

FROM gsoci.azurecr.io/giantswarm/alpine:3.22.2
COPY --from=builder /usr/src/app/frontmatter-validator /usr/local/bin/frontmatter-validator
RUN chmod +x /usr/local/bin/frontmatter-validator

WORKDIR /app

ENTRYPOINT ["/usr/local/bin/frontmatter-validator"]
