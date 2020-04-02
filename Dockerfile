FROM golang:latest as builder

WORKDIR /go/src/github.com/optimizely/agent
COPY . .
RUN make setup build
RUN make ci_build_static_binary

FROM alpine:latest
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/optimizely/agent/bin/optimizely /optimizely
COPY --from=builder /go/src/github.com/optimizely/agent/run_server.sh /run_server.sh
COPY --from=builder /go/src/github.com/optimizely/agent/perfmon_agent.sh /perfmon_agent.sh

RUN apk	update \
	&& apk upgrade \
	&& apk add ca-certificates \
	&& update-ca-certificates \
	&& apk add --update openjdk8-jre tzdata curl unzip bash
	
EXPOSE 4444
CMD /run_server.sh
