FROM golang:1.19.5-buster AS builder

# Build the application
WORKDIR /build
COPY . .
RUN go build -v -o cpu -ldflags "-X main.version=${VERSION}" ./cmd/cpu
RUN go test ./...
WORKDIR /dist
RUN cp /build/cpu ./cpu

# Create the runtime image
FROM scratch
COPY --chown=0:0 --from=builder /dist/cpu /cpu
COPY --chown=0:0 LICENSE /LICENSE
USER 65534
WORKDIR /

LABEL org.opencontainers.image.source=https://github.com/andreax79/go-fcpu/
LABEL org.opencontainers.image.description="fcpu"
LABEL org.opencontainers.image.licenses="Apache-2.0"

ENTRYPOINT ["/cpu"]
