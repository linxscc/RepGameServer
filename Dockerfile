FROM golang:1.24.3 AS build

WORKDIR /src
COPY GoServer/go.mod GoServer/go.sum ./
RUN go mod download

COPY GoServer ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/repgame-server ./main.go

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S repgame \
    && adduser -S -G repgame repgame

COPY --from=build /out/repgame-server /usr/local/bin/repgame-server

USER repgame
EXPOSE 9060
ENTRYPOINT ["/usr/local/bin/repgame-server"]
