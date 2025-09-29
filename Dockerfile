FROM gsoci.azurecr.io/giantswarm/golang:1.25.1

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /usr/local/bin ./...

ENTRYPOINT ["/usr/local/bin/frontmatter-validator"]
