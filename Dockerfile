FROM golang:1.21 as builder
RUN addgroup -u 1000 agentgroup &&\
    useradd -u 1000 agentuser -g agentgroup
WORKDIR /go/src/github.com/optimizely/agent
COPY . .
RUN make setup build &&\
    make ci_build_static_binary

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/optimizely/agent/bin/optimizely /optimizely
COPY --from=builder /etc/passwd /etc/passwd
USER agentuser
CMD ["/optimizely"]