FROM golang:1.25-alpine AS server_builder

WORKDIR /app

RUN apk add --no-cache curl tar xz ca-certificates

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/go/pkg/mod go mod download

RUN curl -L https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz \
  | tar -xJ -C /tmp --strip-components=1

COPY . .

RUN --mount=type=cache,target=/root/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o bin/sgame ./cmd/server

FROM scratch

COPY --from=server_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=server_builder /tmp/ffmpeg /usr/bin/ffmpeg
COPY --from=server_builder /tmp/ffprobe /usr/bin/ffprobe
COPY --from=server_builder /app/bin/sgame /sgame

EXPOSE 80

CMD ["/sgame"]
