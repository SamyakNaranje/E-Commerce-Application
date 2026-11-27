FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/shop .
FROM alpine:3.23
RUN apk add --no-cache ca-certificates && addgroup -S shop && adduser -S -G shop shop
WORKDIR /app
COPY --from=build /out/shop /app/shop
COPY --from=build /src/views /app/views
COPY --from=build /src/assets /app/assets
USER shop
EXPOSE 8080
ENTRYPOINT ["/app/shop"]
