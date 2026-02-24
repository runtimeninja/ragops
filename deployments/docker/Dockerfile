# syntax=docker/dockerfile:1

FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM build AS api
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

FROM build AS worker
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/worker ./cmd/worker

FROM gcr.io/distroless/static-debian12:nonroot AS api_runtime
WORKDIR /
COPY --from=api /bin/api /api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]

FROM gcr.io/distroless/static-debian12:nonroot AS worker_runtime
WORKDIR /
COPY --from=worker /bin/worker /worker
USER nonroot:nonroot
ENTRYPOINT ["/worker"]