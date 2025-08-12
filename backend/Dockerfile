FROM golang:1.24.4 AS server_builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o bin/sgame ./cmd/server

FROM scratch

COPY --from=server_builder /app/bin/sgame .
COPY --from=server_builder /app/.env .

EXPOSE 8080

CMD ["./sgame"]
