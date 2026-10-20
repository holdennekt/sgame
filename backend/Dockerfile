FROM mwader/static-ffmpeg:7.1 AS ffmpeg

FROM golang:1.25-alpine AS server_builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/go/pkg/mod go mod download

COPY . .

RUN --mount=type=cache,target=/root/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o bin/sgame ./cmd/server && \
    CGO_ENABLED=0 go build -o bin/attachment-processor ./cmd/attachment-processor

FROM scratch

COPY --from=server_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=ffmpeg /ffmpeg /usr/bin/ffmpeg
COPY --from=ffmpeg /ffprobe /usr/bin/ffprobe
COPY --from=server_builder /app/bin/sgame /sgame
COPY --from=server_builder /app/bin/attachment-processor /attachment-processor

EXPOSE 80

CMD ["/sgame"]
