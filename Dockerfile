FROM golang:latest

WORKDIR /go/src/github.com/optimizely/agent
COPY . .
RUN make setup build
RUN make ci_build_static_binary

#COPY /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
RUN cp /go/src/github.com/optimizely/agent/bin/optimizely /optimizely
RUN cp /go/src/github.com/optimizely/agent/run_server.sh /run_server.sh

EXPOSE 4444
CMD /run_server.sh
