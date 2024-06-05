FROM golang:1.22 AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" cmd/main.go

FROM alpine:3
WORKDIR /app
COPY --from=build /app/cmd/.env /app
COPY --from=build /app/main /app
ENTRYPOINT ["./main"]