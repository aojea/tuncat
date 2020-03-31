FROM golang:1.13 AS builder

WORKDIR /go/src/tuncat
COPY . .
RUN GOOS=linux CGO_ENABLED=0 go build -o /go/bin/tuncat

# STEP 2: Build Load Balancer Image
FROM alpine:latest
RUN apk add --update iptables conntrack-tools iproute2 curl util-linux \
    && rm -rf /var/cache/apk/*
COPY --from=builder /go/bin/tuncat /bin/tuncat
CMD ["/bin/tuncat","listen","-src-port", "8080"]

# docker run --rm -p 8080:8080 --privileged --sysctl="net.ipv4.ip_forward=1" --sysctl="net.ipv4.conf.all.rp_filter=0" tuncat
