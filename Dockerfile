FROM gsoci.azurecr.io/giantswarm/alpine:3.23.4

ARG TARGETARCH

COPY frontmatter-validator-linux-${TARGETARCH} /usr/local/bin/frontmatter-validator
RUN chmod +x /usr/local/bin/frontmatter-validator

WORKDIR /app

ENTRYPOINT ["/usr/local/bin/frontmatter-validator"]
