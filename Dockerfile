FROM golang:1.24-alpine3.22 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /goflow ./cmd/goflow

FROM gcr.io/distroless/static:nonroot
COPY --from=build /goflow /goflow
COPY migrations /migrations
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/goflow"]
