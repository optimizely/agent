FROM golang:latest

WORKDIR /go/src/github.com/optimizely/agent
COPY . .
RUN make install build
RUN make ci_build_static_binary

COPY /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY /go/src/github.com/optimizely/agent/bin/optimizely /optimizely
COPY /go/src/github.com/optimizely/agent/run_server.sh /run_server.sh

EXPOSE 4444
CMD ./run_server.sh
