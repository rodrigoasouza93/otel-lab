FROM golang:1.22.3 AS build-a
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o serviceA

FROM alpine:3

WORKDIR /app
COPY --from=build-a /app/serviceA .
EXPOSE 8080
CMD ["./serviceA"]