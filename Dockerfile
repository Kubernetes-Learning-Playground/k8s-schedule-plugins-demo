# Build the manager binary
FROM golang:1.20.4-alpine3.18 as builder

WORKDIR /app
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# 坑：
# 报错 go mod download: google.golang.org/api@v0.44.0: read tcp 172.17.0.3:60862->14.204.51.154:443: read: connection reset by peer
# The command '/bin/sh -c go mod download' returned a non-zero code: 1
# make: *** [docker-build] 错误 1
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
# # 需要把该放入的都copy进去，如果报出 package xxxxx is not in GOROOT  => 就是这个问题。
COPY main.go main.go
COPY pkg/ pkg/
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o test-pod-maxNum-scheduler main.go


FROM alpine:3.12
WORKDIR /app
# 需要的文件需要复制过来
COPY --from=builder /app/test-pod-maxNum-scheduler .
USER 65532:65532

ENTRYPOINT ["./test-pod-maxNum-scheduler"]