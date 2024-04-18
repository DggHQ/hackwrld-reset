FROM golang:alpine as builder-base
LABEL builder=true multistage_tag="reset"
RUN apk add --no-cache upx ca-certificates tzdata git

FROM builder-base as builder-modules
LABEL builder=true multistage_tag="reset"
ARG TARGETARCH
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go get github.com/DggHQ/hackwrld-reset/datastore
RUN go get github.com/DggHQ/hackwrld-reset/k8s
RUN go mod download
RUN go mod verify

FROM builder-modules as builder
LABEL builder=true multistage_tag="reset"
ARG TARGETARCH
WORKDIR /build
COPY *.go .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -trimpath -ldflags '-s -w -extldflags="-static"' -v -o reset
RUN upx --best --lzma reset 

FROM alpine:3.17
WORKDIR /app
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /build/reset /usr/bin/
CMD ["reset"]