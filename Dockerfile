FROM golang:alpine as builder
LABEL builder=true multistage_tag="reset"
RUN apk add --no-cache upx ca-certificates tzdata git
ARG TARGETARCH
WORKDIR /build
COPY . .
RUN rm -rf .github Dockerfile
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -trimpath -ldflags '-s -w -extldflags="-static"' -v -o reset
RUN upx --best --lzma reset 

FROM alpine:3.17
WORKDIR /app
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /build/reset /usr/bin/
CMD ["reset"]