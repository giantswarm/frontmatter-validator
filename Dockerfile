FROM gsoci.azurecr.io/giantswarm/golang:1.25.1

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /usr/local/bin ./...

ENTRYPOINT ["/usr/local/bin/frontmatter-validator"]
