# build binary
FROM golang:1.14-alpine AS build

ARG GOOS

ENV CGO_ENABLED=0 \
    GOOS=$GOOS \
    GOARCH=amd64 \
    CGO_CPPFLAGS="-I/usr/include" \
    UID=0 GID=0 \
    CGO_CFLAGS="-I/usr/include" \
    CGO_LDFLAGS="-L/usr/lib -lpthread -lrt -lstdc++ -lm -lc -lgcc -lz " \
    PKG_CONFIG_PATH="/usr/lib/pkgconfig"

RUN apk add --no-cache git make
RUN go get -u golang.org/x/lint/golint
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.30.0


ARG APP_PKG_NAME
WORKDIR /go/src/$APP_PKG_NAME
COPY ./cmd ./cmd
COPY ./pkg ./pkg
COPY ./vendor ./vendor
COPY ./internal ./internal

ARG VERSION=dev
ARG BINARY_NAME

RUN golangci-lint run -E gofmt -E golint -E vet -E goimports
RUN go test -v ./...
RUN go build -v \
    -o /out/service \
    -ldflags "-extldflags "-static" -X main.serviceVersion=$VERSION" \
    ./cmd/$BINARY_NAME

# copy to alpine image
FROM alpine:3.8
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/service /app/service
CMD ["/app/service"]
