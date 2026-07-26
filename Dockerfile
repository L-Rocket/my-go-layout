# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

# ---- runtime stage ----
FROM alpine:3.21
RUN adduser -D -u 10001 appuser
COPY --from=build /out/server /usr/local/bin/server
USER 10001
EXPOSE 8080
ENTRYPOINT ["server"]
