# Build the manager binary
FROM golang:1.14 as builder

WORKDIR /go/src/openkube
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
#RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY vendor/ vendor/
COPY log/ log/
COPY webhook/ webhook/


# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o openkube-manager -mod=vendor main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM alpine:3.8
WORKDIR /
#ENV http_proxy="http://10.49.9.193:1080/"
#ENV https_proxy="http://10.49.9.193:1080/"
# Set the default timezone to Shanghai
# RUN apk update && apk add tzdata \
#     &&  cp -r -f /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime \
#     && echo "Asia/Shanghai" > /etc/timezone \
#     && rm -rf /var/cache/apk/*
#
#ENV http_proxy=""
#ENV https_proxy=""

COPY --from=builder /go/src/openkube/openkube-manager .


#USER nonroot:nonroot
#ENTRYPOINT ["/manager"]
