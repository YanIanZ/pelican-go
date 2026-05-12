FROM golang:1.25.7-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /panel ./cmd/panel

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /panel /panel
ENTRYPOINT ["/panel"]
