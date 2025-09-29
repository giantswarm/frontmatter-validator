FROM gsoci.azurecr.io/giantswarm/golang:1.25.1-alpine3.22 AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v .

FROM gsoci.azurecr.io/giantswarm/alpine:3.22
COPY --from=builder /usr/local/bin/frontmatter-validator /usr/local/bin/frontmatter-validator
RUN chmod +x /usr/local/bin/frontmatter-validator

WORKDIR /app

ENTRYPOINT ["/usr/local/bin/frontmatter-validator"]
